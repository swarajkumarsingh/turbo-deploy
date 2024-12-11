import os from "os";
import fs from "fs";
import dotenv from "dotenv";
import path from "path";
import mime from "mime-types";
import stripAnsi from "strip-ansi";
import { spawn } from "child_process";
import PQueue from "p-queue";
import retry from "async-retry";
import { fileURLToPath } from "url";
import { dirname } from "path";

import { S3Client, PutObjectCommand } from "@aws-sdk/client-s3";
import { SQSClient, SendMessageCommand } from "@aws-sdk/client-sqs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

dotenv.config();

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

const MAX_RETRIES = 3;
const MIN_RETRY_TIMEOUT = 1000;
const MAX_RETRY_TIMEOUT = 5000;

const DEFAULT_BUILD_FOLDER = "build";
const DEFAULT_OUTPUT_FOLDER = "output";
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
  /git\s*.*$/,
  /docker\s*.*$/,
  /rm\s*.*$/,
  /find\s*.*$/,
  /xargs\s*.*$/,
  /bash\s*.*$/,
  /sh\s*.*$/,
  /sudo\s*.*$/,
  /echo\s*.*\$\{.*\}/,
  /python\s*.*$/,
  /perl\s*.*$/,
  /ruby\s*.*$/,
  /node\s*.*$/,
  /make\s*.*$/,
  /tar\s*.*$/,
  /gzip\s*.*$/,
  /bzip2\s*.*$/,
  /xz\s*.*$/,
  /unzip\s*.*$/,
  /unrar\s*.*$/,
  /rar\s*.*$/,
  /dd\s*.*$/,
  /nc\s*.*$/,
  /nmap\s*.*$/,
  /iptables\s*.*$/,
  /ufw\s*.*$/,
  /systemctl\s*.*$/,
  /service\s*.*$/,
  /reboot\s*.*$/,
  /shutdown\s*.*$/,
  /halt\s*.*$/,
  /poweroff\s*.*$/,
  /init\s*.*$/,
  /mount\s*.*$/,
  /umount\s*.*$/,
  /chmod\s*.*$/,
  /chattr\s*.*$/,
  /iptables\s*.*$/,
  /sysctl\s*.*$/,
  /echo\s*.*$/,
  /bc\s*.*$/,
  /tr\s*.*$/,
  /tac\s*.*$/,
  /tee\s*.*$/,
  /awk\s*.*$/,
  /sed\s*.*$/,
  /grep\s*.*$/,
  /cut\s*.*$/,
  /sort\s*.*$/,
  /uniq\s*.*$/,
  /join\s*.*$/,
  /paste\s*.*$/,
  /join\s*.*$/,
  /split\s*.*$/,
  /xargs\s*.*$/,
  /awk\s*.*$/,
  /sed\s*.*$/,
  /rsync\s*.*$/,
  /scp\s*.*$/,
  /wget\s*.*$/,
  /curl\s*.*$/,
  /setuid\s*.*$/,
  /setgid\s*.*$/,
  /chroot\s*.*$/,
  /launchctl\s*.*$/,
  /chkrootkit\s*.*$/,
  /rkhunter\s*.*$/,
  /netstat\s*.*$/,
  /python\s*.*$/,
  /perl\s*.*$/,
  /ruby\s*.*$/,
  /npm\s*.*$/,
  /yarn\s*.*$/,
  /pip\s*.*$/,
  /gem\s*.*$/,
  /composer\s*.*$/,
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

const queue = new PQueue({ concurrency: 5 });

async function init() {
  try {
    console.log("Executing script.js");
    await publishLog({ message: `Running Environment: ${ENVIRONMENT}` });
    await publishLog({ message: "Build Started..." });
    await pushDeploymentStatus(DeploymentStatus.PROG);

    const outDirPath = sanitizePath(
      path.join(__dirname, DEFAULT_OUTPUT_FOLDER)
    );

    const validBuildCommand = checkValidBuildCommandFromPackageFile(outDirPath);
    if (!validBuildCommand) {
      await publishLog({ message: "Vulnerable build command found :(" });
      await pushDeploymentStatus(DeploymentStatus.FAIL);
      process.exit(1);
    }

    await publishLog({ message: "Executing npm install & npm build commands" });

    const p = spawn("npm", ["install", "&&", "npm", "run", "build"], {
      cwd: outDirPath,
      shell: true,
    });

    p.stdout.on("data", async function (data) {
      await publishLog({ message: data.toString() });
    });

    p.stderr.on("data", async function (data) {
      await publishLog({
        message: data.toString(),
        logType: LogType.ERROR,
      });
    });

    p.on("error", async function (error) {
      await publishLog({
        message: `Error starting process: ${error.message}`,
        logType: LogType.ERROR,
        cause: error.cause,
        name: error.name,
        stack: error.stack.toString(),
      });
      process.exit(1);
    });

    p.on("exit", async function (code) {
      if (code !== 0) {
        await publishLog({
          message: `Build process exited with code ${code}`,
          logType: LogType.ERROR,
        });
        await pushDeploymentStatus(DeploymentStatus.FAIL);
        process.exit(1);
      }
    });

    p.on("close", async function () {
      await publishLog({
        message: "npm install & npm build completed successfully",
      });

      const outputFolder =
        outputFolders.find((folder) =>
          fs.existsSync(path.join(__dirname, DEFAULT_OUTPUT_FOLDER, folder))
        ) || DEFAULT_BUILD_FOLDER;

      const distFolderPath = path.join(
        __dirname,
        DEFAULT_OUTPUT_FOLDER,
        outputFolder
      );
      if (!fs.existsSync(distFolderPath)) {
        throw new Error(`Output folder "${outputFolder}" does not exist.`);
      }

      const filesToUpload = getAllFiles(distFolderPath);
      if (filesToUpload.length === 0) {
        throw new Error(
          `No files found in the output folder: ${distFolderPath}`
        );
      }

      await publishLog({
        message: `Starting to upload in dir ${outputFolder}`,
      });

      for (const filePath of filesToUpload) {
        queue.add(async () => {
          const relativeFilePath = path.relative(distFolderPath, filePath);
          const s3Key = `__outputs/${DEPLOYMENT_ID}/${relativeFilePath.replace(
            /\\/g,
            "/"
          )}`;

          await publishLog({ message: `Uploading ${relativeFilePath}` });

          const command = new PutObjectCommand({
            Bucket: S3_BUCKET_NAME,
            Key: s3Key,
            Body: fs.createReadStream(filePath),
            ContentType: mime.lookup(filePath) || "application/octet-stream",
          });

          try {
            await retry(
              async () => {
                await s3Client.send(command);
                await publishLog({ message: `Uploaded ${relativeFilePath}` });
              },
              {
                retries: MAX_RETRIES,
                factor: 2,
                minTimeout: MIN_RETRY_TIMEOUT,
                maxTimeout: MAX_RETRY_TIMEOUT,
                onRetry: async (error, attempt) => {
                  console.log(
                    `Retry attempt ${attempt} failed for ${relativeFilePath}: ${error.message}`
                  );
                  await publishLog({
                    message: `Retry attempt ${attempt} failed for ${relativeFilePath}: ${error.message}`,
                    logType: LogType.ERROR,
                  });
                },
              }
            );
          } catch (uploadError) {
            console.error(
              `Failed to upload file ${relativeFilePath}: ${uploadError.message}`
            );
            await publishLog({
              message: `Failed to upload file ${relativeFilePath}: ${uploadError.message}`,
              logType: LogType.ERROR,
            });
            throw new Error("Failed to upload file ${relativeFilePath}: ${uploadError.message}"); 
          }
        });
      }

      await queue.onIdle();

      await publishLog({ message: "All files uploaded successfully." });
      await pushDeploymentStatus(DeploymentStatus.READY);
      await publishLog({ message: "Done..." });
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
