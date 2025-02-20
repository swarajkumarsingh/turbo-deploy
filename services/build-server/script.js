import os from "os";
import fs from "fs";
import http from "http";
import path from "path";
import { dirname } from "path";
import { fileURLToPath } from "url";
import { spawn } from "child_process";
import { VULNERABLE_COMMANDS } from "./vulnerableCommands.js";

import https from "https";
import dotenv from "dotenv";
import PQueue from "p-queue";
import mime from "mime-types";
import retry from "async-retry";
import stripAnsi from "strip-ansi";
import { S3Client, PutObjectCommand } from "@aws-sdk/client-s3";
import { SQSClient, SendMessageCommand } from "@aws-sdk/client-sqs";

dotenv.config();

const DeploymentStatus = {
  PROG: "PROG",
  FAIL: "FAIL",
  READY: "READY",
};

const LogType = {
  INFO: "INFO",
  WARN: "WARN",
  ERROR: "ERROR",
};

const {
  APP_NAME,
  AWS_REGION,
  PROJECT_ID,
  ENVIRONMENT,
  DEPLOYMENT_ID,
  LOG_QUEUE_URL,
  S3_BUCKET_NAME,
  BUILD_TEST_URL,
  EMAIL_QUEUE_URL,
  RECIPIENT_EMAIL,
  STATUS_QUEUE_URL,
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

const MAX_RETRIES = 3;
const MIN_RETRY_TIMEOUT = 1000;
const MAX_RETRY_TIMEOUT = 5000;

const APPLICATION_STATUS_SUBJECT =
  "Application Deployment Status | Turbo-Deploy";
const SUCCESS_DEPLOYMENT_BODY = `Congrats! Your application with Project id: ${PROJECT_ID} & Deployment id: ${DEPLOYMENT_ID} is deployed successfully. Please Visit Turbo Deploy for more details`;

const DEFAULT_BUILD_FOLDER = "build";
const DEFAULT_OUTPUT_FOLDER = "output";
const outputFolders = ["dist", "build", "public", "release"];

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const MAX_FILE_SIZE = 1024 * 1024;
const PACKAGE_JSON_PATH = path.join(
  __dirname,
  DEFAULT_OUTPUT_FOLDER,
  "package.json"
);

const queue = new PQueue({ concurrency: 5 });

const requiredEnvVars = [
  "APP_NAME",
  "AWS_REGION",
  "PROJECT_ID",
  "ENVIRONMENT",
  "LOG_QUEUE_URL",
  "DEPLOYMENT_ID",
  "S3_BUCKET_NAME",
  "RECIPIENT_EMAIL",
  "EMAIL_QUEUE_URL",
  "STATUS_QUEUE_URL",
  "AWS_ACCESS_KEY_ID",
  "AWS_SECRET_ACCESS_KEY",
];

requiredEnvVars.forEach(async (varName) => {
  if (!process.env[varName] || process.env[varName].trim() === "") {
    console.error(`ERROR: Missing or empty environment variable: ${varName}`);
    await pushDeploymentStatus(DeploymentStatus.FAIL);
    process.exit(1);
  }
});

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
  message,
  logType = LogType.INFO,
  cause = "",
  name = "",
  stack = "",
}) {
  const host = os.hostname();

  if (!Object.values(LogType).includes(logType) || logType.trim() === "") {
    logType = LogType.INFO;
  }

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

  if (!obj.message || obj.message.trim() === "") {
    console.log("Skipping log with no message");
    return;
  }

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
    status: sanitizedStatus,
    host,
    timestamp: new Date().toISOString(),
  };
  await publishToQueue(STATUS_QUEUE_URL, statusMessage);
  console.log(`Status: ${status}`);
}

async function pushEmailQueue(recipient_email, subject, body) {
  const emailMessage = {
    appName: APP_NAME,
    recipient_email: recipient_email || RECIPIENT_EMAIL,
    subject: subject || APPLICATION_STATUS_SUBJECT,
    body: body || SUCCESS_DEPLOYMENT_BODY,
    projectId: PROJECT_ID,
    deploymentId: DEPLOYMENT_ID,
    timestamp: new Date().toISOString(),
  };
  await publishToQueue(EMAIL_QUEUE_URL, emailMessage);
  console.log(`Email Status: Success`);
}

function sanitizePath(dirPath) {
  if (!path.isAbsolute(dirPath)) throw new Error("Path must be absolute");
  const sanitizedPath = path.normalize(dirPath);
  if (sanitizedPath.includes(".."))
    throw new Error("Path must not navigate outside allowed directories");
  return sanitizedPath;
}

async function isStatusCode200(url) {
  return new Promise((resolve, _) => {
    try {
      if (!url) {
        throw new Error("The URL environment variable is not defined.");
      }

      const protocol = url.startsWith("https") ? https : http;

      protocol
        .get(url, (res) => {
          resolve(res.statusCode === 200);
        })
        .on("error", (error) => {
          console.error(
            "Warning; something went wrong occurred while checking the status code:",
            error.message
          );
          resolve(false);
        });
    } catch (error) {
      console.error("Error occurred:", error.message);
      resolve(false);
    }
  });
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

function getDeploymentSuccessHtml(projectId, deploymentId) {
  return `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Application Deployment Status | Turbo-Deploy</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                margin: 0;
                padding: 0;
                background-color: #f9f9f9;
                color: #333;
                line-height: 1.6;
            }
            .container {
                max-width: 600px;
                margin: 20px auto;
                background: #ffffff;
                border-radius: 8px;
                box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
                padding: 20px;
                border: 1px solid #ddd;
            }
            h1 {
                color: #4CAF50;
                text-align: center;
                font-size: 24px;
            }
            p {
                margin: 10px 0;
            }
            .details {
                background: #f1f1f1;
                border: 1px solid #ddd;
                padding: 10px;
                border-radius: 6px;
                margin: 20px 0;
            }
            a {
                color: #4CAF50;
                text-decoration: none;
            }
            a:hover {
                text-decoration: underline;
            }
            .footer {
                text-align: center;
                margin-top: 20px;
                font-size: 12px;
                color: #666;
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>Application Deployment Status</h1>
            <p>Dear User,</p>
            <p>Congratulations! Your application has been successfully deployed. Below are the deployment details:</p>
            <div class="details">
                <p><strong>Project ID:</strong> ${projectId}</p>
                <p><strong>Deployment ID:</strong> ${deploymentId}</p>
            </div>
            <p>You can now access your deployed application and explore its features. For further details, please visit 
                <a href="#" target="_blank">Turbo Deploy</a>.
            </p>
            <p>Thank you for choosing Turbo-Deploy for your application deployment needs. We look forward to supporting your future projects.</p>
            <div class="footer">
                <p>&copy; 2024 Turbo-Deploy. All rights reserved.</p>
            </div>
        </div>
    </body>
    </html>
    `;
}

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
        message: `Error starting process: ${error}`,
        logType: LogType.ERROR,
        cause: error.cause,
        name: error.name,
        stack: error.stack.toString(),
      });
      await pushDeploymentStatus(DeploymentStatus.FAIL);
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
            throw new Error(
              "Failed to upload file ${relativeFilePath}: ${uploadError.message}"
            );
          }
        });
      }

      await queue.onIdle();

      await publishLog({ message: "All files uploaded successfully." });

      const ok = isStatusCode200(BUILD_TEST_URL);
      if (!ok) {
        await publishLog({
          message: `Project deployed but in undefined state. Try again later. URL: ${BUILD_TEST_URL}`,
          logType: LogType.WARN,
        });
      }

      await publishLog({ message: "Deployment testing completed :)" });
      await pushDeploymentStatus(DeploymentStatus.READY);
      await pushEmailQueue(
        RECIPIENT_EMAIL,
        APPLICATION_STATUS_SUBJECT,
        getDeploymentSuccessHtml(PROJECT_ID, DEPLOYMENT_ID),
      );
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
