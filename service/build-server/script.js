require("dotenv").config();
const fs = require("fs");
const path = require("path");
const mime = require("mime-types");
const { exec, execFile } = require("child_process");

const { S3Client, PutObjectCommand } = require("@aws-sdk/client-s3");
const { SQSClient, SendMessageCommand } = require("@aws-sdk/client-sqs");

const requiredEnvVars = [
  "PROJECT_ID",
  "DEPLOYMENT_ID",
  "LOG_QUEUE_URL",
  "STATUS_QUEUE_URL",
  "S3_BUCKET_NAME",
  "AWS_REGION",
  "AWS_ACCESS_KEY_ID",
  "AWS_SECRET_ACCESS_KEY",
];

requiredEnvVars.forEach((varName) => {
  if (!process.env[varName]) {
    console.error(`Missing environment variable: ${varName}`);
    process.exit(1);
  }
});

const {
  PROJECT_ID,
  DEPLOYMENT_ID,
  LOG_QUEUE_URL,
  STATUS_QUEUE_URL,
  S3_BUCKET_NAME,
  AWS_REGION,
  AWS_ACCESS_KEY_ID,
  AWS_SECRET_ACCESS_KEY,
} = process.env;

const s3Client = new S3Client({
  region: AWS_REGION,
  credentials: {
    accessKeyId: AWS_ACCESS_KEY_ID,
    secretAccessKey: AWS_SECRET_ACCESS_KEY,
  },
});

const sqsClient = new SQSClient({
  region: AWS_REGION,
  credentials: {
    accessKeyId: AWS_ACCESS_KEY_ID,
    secretAccessKey: AWS_SECRET_ACCESS_KEY,
  },
});

const DeploymentStatus = {
  PROG: "PROG",
  FAIL: "FAIL",
  READY: "READY",
};

const EXEC_TIMEOUT = 600000; // 10 minutes
const MAX_BUFFER = 1024 * 1024 * 10; // 10MB
const MAX_RETRIES = 3;

/**
 * Utility: Publish a message to an SQS queue.
 * @param {string} queueUrl - The SQS queue URL.
 * @param {Object} message - The message payload.
 */
async function publishToQueue(queueUrl, message) {
  try {
    const command = new SendMessageCommand({
      QueueUrl: queueUrl,
      MessageBody: JSON.stringify(message),
    });
    await sqsClient.send(command);
  } catch (error) {
    console.error("Error publishing to SQS:", error);
  }
}

/**
 * Logs a message to the log queue.
 * @param {string} log - The log message to publish.
 */
async function publishLog(log) {
  const logMessage = {
    projectId: PROJECT_ID,
    deploymentId: DEPLOYMENT_ID,
    log,
    timestamp: new Date().toISOString(),
  };
  await publishToQueue(LOG_QUEUE_URL, logMessage);
  console.log(`Log: ${log}`);
}

/**
 * Pushes deployment status to the status queue.
 * @param {string} status - Deployment status (PROG, READY, FAIL).
 */
async function pushDeploymentStatus(status) {
  const statusMessage = {
    projectId: PROJECT_ID,
    deploymentId: DEPLOYMENT_ID,
    status,
    timestamp: new Date().toISOString(),
  };
  await publishToQueue(STATUS_QUEUE_URL, statusMessage);
  console.log(`Status: ${status}`);
}

/**
 * Validates and sanitizes a directory path.
 * @param {string} dirPath - The path to validate.
 * @returns {string} - The sanitized path.
 */
function sanitizePath(dirPath) {
  if (!path.isAbsolute(dirPath)) throw new Error("Path must be absolute");
  const sanitizedPath = path.normalize(dirPath);
  if (sanitizedPath.includes(".."))
    throw new Error("Path must not navigate outside allowed directories");
  return sanitizedPath;
}

// List of potentially vulnerable and dangerous npm commands
const VULNERABLE_NPM_COMMANDS = [
  // System and shell commands
  "install",
  "build",
  "run build",
  "uninstall",
  "update",
  "run",
  "start",
  "test",

  // Potentially risky operations
  "publish",
  "link",
  "pack",
  "rebuild",
  "audit",
  "config",

  // Script execution commands
  "run-script",
  "explore",
  "doctor",
  "init",

  // Advanced package management
  "version",
  "prune",
  "shrinkwrap",
];

// List of explicitly blocked commands
const BLOCKED_COMMANDS = [
  // Commands that could potentially expose or modify system resources
  "rm",
  "del",
  "delete",
  "curl",
  "wget",
  "cat",
  "mv",
  "cp",
  "ssh",
  "scp",
  "exec",
  "eval",
  "sudo",
  "chmod",
  "chown",

  // Potentially destructive npm script names
  "destroy",
  "wipe",
  "nuke",
  "remove-all",
];

/**
 * Enhanced command validation with multiple security checks
 * @param {string} command - The npm command to validate
 * @throws {Error} - If the command is considered unsafe
 */
/**
 * Enhanced command validation with multiple security checks
 * @param {string} command - The npm command to validate
 * @throws {Error} - If the command is considered unsafe
 */
function validateNpmCommand(command) {
  // Check if command is in the allowed list
  if (!VULNERABLE_NPM_COMMANDS.includes(command)) {
    throw new Error(`Unauthorized npm command: ${command}`);
  }

  // Additional check for blocked command patterns
  const lowercaseCommand = command.toLowerCase();
  const hasBlockedCommand = BLOCKED_COMMANDS.some((blockedCmd) =>
    lowercaseCommand.includes(blockedCmd)
  );

  if (hasBlockedCommand) {
    throw new Error(`Potentially dangerous command detected: ${command}`);
  }

  const sanitizedCommand = command.replace(/[^a-zA-Z0-9-\s]/g, "");
  if (sanitizedCommand !== command) {
    throw new Error("Invalid characters in npm command");
  }

  return sanitizedCommand;
}

/**
 * Secure NPM command execution function
 * @param {string} command - The npm command to run
 * @param {string} cwd - Current working directory
 * @returns {Promise} - Resolves with command output
 */
function runSecureNpmCommand(command, cwd) {
  return new Promise((resolve, reject) => {
    try {
      // Validate the command before execution
      const sanitizedCommand = validateNpmCommand(command);

      // Use execFile for secure execution
      const npmProcess = execFile(
        "npm",
        [sanitizedCommand],
        {
          cwd: cwd,
          shell: false,
          env: {
            PATH: process.env.PATH,
            HOME: process.env.HOME,
            NODE_ENV: "production",
          },
          timeout: EXEC_TIMEOUT,
          maxBuffer: MAX_BUFFER,
        },
        (error, stdout, stderr) => {
          if (error) {
            console.error(`Secure npm ${command} error:`, error);
            console.log(stderr);
            reject(error);
          } else {
            console.log(`Secure npm ${command} output:`, stdout);
            resolve(stdout);
          }
        }
      );
    } catch (validationError) {
      reject(validationError);
    }
  });
}

/**
 * Uploads a file to S3 with retries for transient errors.
 * @param {string} filePath - The file path to upload.
 * @param {string} s3Key - The S3 key.
 */
async function uploadFileToS3(filePath, s3Key) {
  for (let attempt = 1; attempt <= MAX_RETRIES; attempt++) {
    try {
      const command = new PutObjectCommand({
        Bucket: S3_BUCKET_NAME,
        Key: s3Key,
        Body: fs.createReadStream(filePath),
        ContentType: mime.lookup(filePath) || "application/octet-stream",
      });
      await s3Client.send(command);
      console.log(`Uploaded: ${filePath}`);
      return;
    } catch (error) {
      console.error(`Attempt ${attempt}: Failed to upload ${filePath}:`, error);
      if (attempt === MAX_RETRIES) throw error;
      await new Promise((resolve) => setTimeout(resolve, 1000 * attempt));
    }
  }
}

// Recursive function to get all files from a directory
const getAllFiles = (dirPath, files = []) => {
  const items = fs.readdirSync(dirPath, { withFileTypes: true });
  for (const item of items) {
    const itemPath = path.join(dirPath, item.name);
    if (item.isDirectory()) {
      // Recursively process directories
      getAllFiles(itemPath, files);
    } else {
      files.push(itemPath); // Add file to the list
    }
  }
  return files;
};

// TODO react package.json and check for bad build command
function checkValidBuildCommandFromPackageFile(path) {
  return true;
}

async function init() {
  console.log("Executing script.js");
  await publishLog("Build Started...");
  await pushDeploymentStatus(DeploymentStatus.PROG);

  const outDirPath = sanitizePath(path.join(__dirname, "output"));
  const validBuildCommand = checkValidBuildCommandFromPackageFile(outDirPath);
  if (!validBuildCommand) {
    await publishLog("Vulnerable build command found :(");
    await pushDeploymentStatus(DeploymentStatus.FAIL); 
  }

  try {
    // Secure npm install
    await publishLog("Executing npm install command");
    // await runSecureNpmCommand("install", outDirPath);
    await publishLog("npm install completed successfully");

    // Secure npm build
    await publishLog("Executing npm build command");
    
    // await runSecureNpmCommand("run build", outDirPath);
    const p = exec(`cd ${outDirPath} && npm install && npm run build`);

    p.stdout.on("data", async function (data) {
      console.log(data.toString());
      publishLog(data.toString());
    });

    p.stdout.on("error", async function (data) {
      console.log("Error", data.toString());
      await publishLog(`error: ${data.toString()}`);
   process.exit(1);
    });

    p.stdout.on("close", async function () {
      await publishLog("npm build completed successfully");

    const outputFolder = fs.existsSync(path.join(__dirname, "output", "dist"))
      ? "dist"
      : "build";
     const distFolderPath = path.join(__dirname, "output", outputFolder);
     const filesToUpload = getAllFiles(distFolderPath); 

     await publishLog(`Starting to upload`);

     for (const filePath of filesToUpload) {
       const relativeFilePath = path.relative(distFolderPath, filePath);
       const s3Key = `__outputs/${DEPLOYMENT_ID}/${relativeFilePath.replace(
         /\\/g,
         "/"
       )}`;

       console.log("uploading", relativeFilePath);
       await publishLog(`uploading ${relativeFilePath}`);

       const command = new PutObjectCommand({
         Bucket: S3_BUCKET_NAME,
         Key: s3Key,
         Body: fs.createReadStream(filePath),
         ContentType: mime.lookup(filePath) || "application/octet-stream",
       });

       await s3Client.send(command);
       await publishLog(`uploaded ${relativeFilePath}`);
       console.log("uploaded", relativeFilePath);
     }

     await publishLog(`Done`);
     console.log("Done...");
     process.exit(0);

    });
  } catch (error) {
    console.error("Deployment failed:", error);
    await publishLog(`Deployment failed: ${error.message}`);
    await pushDeploymentStatus(DeploymentStatus.FAIL);
   process.exit(1);
  }
}

process.on("SIGTERM", async () => {
  console.log("SIGTERM received, shutting down gracefully");
  process.exit(0);
});

process.on("unhandledRejection", async (reason, promise) => {
  console.log("Unhandled Rejection at:", promise, "reason:", reason);
});

init();
