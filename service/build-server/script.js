const os = require("os");
const fs = require("fs");
require("dotenv").config();
const path = require("path");
const mime = require("mime-types");
const stripAnsi = require("strip-ansi");
const { exec } = require("child_process");

const { S3Client, PutObjectCommand } = require("@aws-sdk/client-s3");
const { SQSClient, SendMessageCommand } = require("@aws-sdk/client-sqs");

const requiredEnvVars = [
  "APP_NAME",
  "ENVIRONMENT",
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
  APP_NAME,
  ENVIRONMENT,
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

const LogType = {
  INFO: "INFO",
  ERROR: "ERROR",
  WARN: "WARN",
};

let outputFolder = "build";
const outputFolders = ["dist", "build", "public", "release"];

const MAX_FILE_SIZE = 1024 * 1024;
const PACKAGE_JSON_PATH = path.join(__dirname, "output", "package.json");
const VULNERABLE_COMMANDS = [
  /rm\s*-rf\s*$/,
  /curl\s*http[s]?:\/\//,
  /wget\s*http[s]?:\/\//,
  /eval\s*\(.*\)/,
  /exec\s*\(.*\)/,
  /\$\(.*\)/,
  /cat\s*.*$/,
  /more\s*.*$/,
  /tail\s*-f\s*.*$/,
  /head\s*-n\s*\d*\s*.*$/,
  /less\s*.*$/,
  /touch\s*.*$/,
  /chmod\s*.*$/,
  /chown\s*.*$/,
  /ln\s*.*$/,
  /env\s*$/,
  /printenv\s*$/,
  /export\s*.*$/,
  /ps\s*-aux$/,
  /top\s*$/,
  /pstree\s*$/,
  /kill\s*$/,
  /killall\s*$/,
  /lsof\s*$/,
  /netstat\s*$/,
  /scp\s*.*$/,
  /ssh\s*.*$/,
];

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

async function publishLog({
  message = "",
  logType = LogType.INFO,
  cause = "",
  name = "",
  stack = "",
}) {
  const host = os.hostname();
  const obj = {
    appName: APP_NAME,
    message: stripAnsi(message),
    logType,
    projectId: PROJECT_ID,
    deploymentId: DEPLOYMENT_ID,
    environment: process.env.ENVIRONMENT || "DEV",
    host,
    cause: stripAnsi(cause),
    name: stripAnsi(name),
    stack: stripAnsi(stack),
    timestamp: new Date().toISOString(),
  };

  await publishToQueue(LOG_QUEUE_URL, obj);
  console.log(`Log: ${message}`);
}

async function pushDeploymentStatus(status) {
  const host = os.hostname();
  const sanitizedStatus = stripAnsi(status);

  const statusMessage = {
    app_name: APP_NAME,
    projectId: PROJECT_ID,
    deploymentId: DEPLOYMENT_ID,
    sanitizedStatus,
    host,
    timestamp: new Date().toISOString(),
  };
  await publishToQueue(STATUS_QUEUE_URL, statusMessage);
  console.log(`Status: ${status}`);
}

function sanitizePath(dirPath) {
  if (!path.isAbsolute(dirPath)) throw new Error("Path must be absolute");
  const sanitizedPath = path.normalize(dirPath);
  if (sanitizedPath.includes(".."))
    throw new Error("Path must not navigate outside allowed directories");
  return sanitizedPath;
}

const getAllFiles = (dirPath, files = []) => {
  const items = fs.readdirSync(dirPath, { withFileTypes: true });
  for (const item of items) {
    const itemPath = path.join(dirPath, item.name);
    if (item.isDirectory()) getAllFiles(itemPath, files);
    else files.push(itemPath);
  }
  return files;
};

function checkValidBuildCommandFromPackageFile() {
  try {
    const stats = fs.statSync(PACKAGE_JSON_PATH);
    if (stats.size > MAX_FILE_SIZE) return false;

    const data = fs.readFileSync(PACKAGE_JSON_PATH, {
      encoding: "utf-8",
      flag: "r",
    });

    let packageJson;
    try {
      packageJson = JSON.parse(data);
    } catch {
      return false;
    }

    if (
      !packageJson.scripts ||
      !packageJson.scripts.build ||
      typeof packageJson.scripts.build !== "string"
    )
      return false;

    const buildCommand = packageJson.scripts.build;
    const sanitizedCommand = buildCommand.trim().replace(/^["']|["']$/g, "");

    return !VULNERABLE_COMMANDS.some((pattern) =>
      pattern.test(sanitizedCommand)
    );
  } catch {
    return false;
  }
}

async function init() {
  try {
    console.log("Executing script.js");
    await publishLog({ message: `Running Environment: ${ENVIRONMENT}` });
    await publishLog({ message: "Build Started..." });
    await pushDeploymentStatus(DeploymentStatus.PROG);

    const outDirPath = sanitizePath(path.join(__dirname, "output"));

    const validBuildCommand = checkValidBuildCommandFromPackageFile(outDirPath);
    if (!validBuildCommand) {
      await publishLog({ message: "Vulnerable build command found :(" });
      await pushDeploymentStatus(DeploymentStatus.FAIL);
      process.exit(1);
    }

    await publishLog({ message: "Executing npm install & npm build commands" });
    const p = exec(`cd ${outDirPath} && npm install && npm run build`);

    p.stdout.on("data", async function (data) {
      await publishLog({ message: data.toString() });
    });

    p.stdout.on("error", async function (error) {
      await publishLog({
        message: `error: ${error.message}`,
        logType: LogType.ERROR,
        cause: error.cause,
        name: error.name,
        stack: error.stack.toString(),
      });
      process.exit(1);
    });

    p.stdout.on("close", async function () {
      await publishLog({
        message: "npm install & npm build completed successfully",
      });

      for (const folder of outputFolders) {
        if (fs.existsSync(path.join(__dirname, "output", folder))) {
          outputFolder = folder;
          break;
        }
      }

      const distFolderPath = path.join(__dirname, "output", outputFolder);
      const filesToUpload = getAllFiles(distFolderPath);

      await publishLog({
        message: `Starting to uploading in dir ${outputFolder}`,
      });

      for (const filePath of filesToUpload) {
        const relativeFilePath = path.relative(distFolderPath, filePath);
        const s3Key = `__outputs/${DEPLOYMENT_ID}/${relativeFilePath.replace(
          /\\/g,
          "/"
        )}`;

        await publishLog({ message: `uploading ${relativeFilePath}` });

        const command = new PutObjectCommand({
          Bucket: S3_BUCKET_NAME,
          Key: s3Key,
          Body: fs.createReadStream(filePath),
          ContentType: mime.lookup(filePath) || "application/octet-stream",
        });

        await s3Client.send(command);
        await publishLog({ message: `uploaded ${relativeFilePath}` });
      }

      await publishLog({ message: `Done...` });
      await pushDeploymentStatus(DeploymentStatus.READY);
      process.exit(0);
    });
  } catch (error) {
    console.error("Deployment failed:", error);
    await publishLog({
      message: `Deployment failed: ${error.message}`,
      logType: LogType.ERROR,
    });
    await pushDeploymentStatus(DeploymentStatus.FAIL);
    process.exit(1);
  }
}

process.on("SIGTERM", async () => {
  console.log("SIGTERM received, shutting down gracefully");
  process.exit(0);
});

process.on("unhandledRejection", async (reason, promise) => {
  console.error("Unhandled Rejection at:", promise, "reason:", reason);
  await publishLog({
    message: `Unhandled Rejection: ${reason}`,
    logType: LogType.ERROR,
  });
  await pushDeploymentStatus(DeploymentStatus.FAIL);
  process.exit(1);
});


init();
