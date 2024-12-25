require("dotenv").config();
const path = require("path");
const cors = require("cors");
const Queue = require("bull");
const { Pool } = require("pg");
const Redis = require("ioredis");
const morgan = require("morgan");
const helmet = require("helmet");
const express = require("express");
const winston = require("winston");
const httpProxy = require("http-proxy");
const promClient = require("prom-client");
const compression = require("compression");
const rateLimit = require("express-rate-limit");
const { createBullBoard } = require("@bull-board/api");
const { ExpressAdapter } = require("@bull-board/express");
const { BullAdapter } = require("@bull-board/api/bullAdapter");

// Environment variables validation with schema
const requiredEnvVars = {
  PORT: { type: "number", default: 8000 },
  S3_BASE_PATH: { type: "string", required: true },
  DB_PORT: { type: "number", required: true },
  DB_USER: { type: "string", required: true },
  DB_DATABASE: { type: "string", required: true },
  DB_PASSWORD: { type: "string", required: true },
  DB_HOST: { type: "string", required: true },
  REDIS_HOST: { type: "string", required: true },
  REDIS_PORT: { type: "number", required: true },
  REDIS_PASSWORD: { type: "string", required: true },
  NODE_ENV: { type: "string", default: "development" },
};

const validateEnv = () => {
  const errors = [];
  for (const [key, config] of Object.entries(requiredEnvVars)) {
    const value = process.env[key];
    if (config.required && (!value || value.trim() === "")) {
      errors.push(`Missing required environment variable: ${key}`);
    } else if (value && config.type === "number" && isNaN(Number(value))) {
      errors.push(`Invalid number format for environment variable: ${key}`);
    }
  }
  if (errors.length > 0) {
    logger.error("Environment validation failed", { errors });
    process.exit(1);
  }
};

validateEnv();

const S3_BASE_PATH = process.env.S3_BASE_PATH;

// Prometheus metrics setup
const register = new promClient.Registry();
promClient.collectDefaultMetrics({ register });

const httpRequestDuration = new promClient.Histogram({
  name: "http_request_duration_seconds",
  help: "Duration of HTTP requests in seconds",
  labelNames: ["method", "route", "status_code"],
  buckets: [0.1, 0.5, 1, 2, 5],
});
register.registerMetric(httpRequestDuration);

// Queue setup for background jobs
const deploymentQueue = new Queue("deployment-processing", {
  redis: {
    host: process.env.REDIS_HOST,
    port: process.env.REDIS_PORT,
    password: process.env.REDIS_PASSWORD,
  },
});

// Bull Board setup for queue monitoring
const serverAdapter = new ExpressAdapter();
createBullBoard({
  queues: [new BullAdapter(deploymentQueue)],
  serverAdapter,
});

const logger = winston.createLogger({
  level: process.env.LOG_LEVEL || "info",
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.errors({ stack: true }),
    winston.format.metadata({ fillExcept: ["message", "level", "timestamp"] }), // Include all metadata except these fields
    winston.format.printf(({ timestamp, level, message, metadata }) => {
      const meta = Object.keys(metadata).length ? JSON.stringify(metadata) : "";
      return `${timestamp} ${level}: ${message} ${meta}`;
    })
  ),
  defaultMeta: { service: "proxy-service" },
  transports: [
    new winston.transports.Console({
      format: winston.format.combine(
        winston.format.colorize(),
        winston.format.printf(({ timestamp, level, message, metadata }) => {
          const meta = Object.keys(metadata).length
            ? JSON.stringify(metadata)
            : "";
          return `${timestamp} ${level}: ${message} ${meta}`;
        })
      ),
    }),
    new winston.transports.File({
      filename: path.join(process.env.LOG_DIR || "logs", "error.log"),
      level: "error",
      maxsize: 5242880,
      maxFiles: 5,
    }),
    new winston.transports.File({
      filename: path.join(process.env.LOG_DIR || "logs", "combined.log"),
      maxsize: 5242880,
      maxFiles: 5,
    }),
  ],
});

// Redis client setup
const redisConfig = {
  host: process.env.REDIS_HOST,
  port: process.env.REDIS_PORT,
  password: process.env.REDIS_PASSWORD,
  retryStrategy: (times) => {
    const delay = Math.min(times * 50, 2000);
    return delay;
  },
  showFriendlyErrorStack: true,
};

const redis = new Redis(redisConfig);

redis.on("connect", () => {
  logger.info("connected to redis");
});

redis.on("error", (err) => {
  logger.error("Redis error", { error: err });
});

// Database connection with connection pooling and retry logic
const createPool = async () => {
  const pool = new Pool({
    host: process.env.DB_HOST,
    port: process.env.DB_PORT,
    user: process.env.DB_USER,
    database: process.env.DB_DATABASE,
    password: process.env.DB_PASSWORD,
    ssl:
      process.env.NODE_ENV === "production"
        ? { rejectUnauthorized: true }
        : { rejectUnauthorized: false },
    max: 20,
    idleTimeoutMillis: 30000,
    connectionTimeoutMillis: 2000,
    application_name: "proxy-service",
  });

  try {
    await pool.query("SELECT NOW()");
    logger.info("Database connection established");
    return pool;
  } catch (err) {
    logger.error("Failed to connect to database", { error: err });
    throw err;
  }
};

const limiter = rateLimit({
  store: redis,
  windowMs: 15 * 60 * 1000,
  max: process.env.NODE_ENV === "production" ? 100 : 1000,
  standardHeaders: true,
  legacyHeaders: false,
  handler: (req, res) => {
    logger.warn("Rate limit exceeded", {
      ip: req.ip,
      path: req.path,
    });
    res.status(429).json({
      error: "Too many requests, please try again later.",
    });
  },
});

const app = express();

if (process.env.NODE_ENV === "development") {
  app.use(morgan("dev"));
} else {
  app.use(morgan("combined", { stream: winston.stream.write }));
}

app.use(
  helmet({
    contentSecurityPolicy: {
      directives: {
        defaultSrc: ["'self'"],
        styleSrc: ["'self'", "'unsafe-inline'"],
        imgSrc: ["'self'", "data:", "https:"],
        scriptSrc: ["'self'"],
        upgradeInsecureRequests:
          process.env.NODE_ENV === "production" ? [] : null,
      },
    },
    crossOriginEmbedderPolicy: true,
    crossOriginOpenerPolicy: true,
    crossOriginResourcePolicy: true,
    hsts: process.env.NODE_ENV === "production",
  })
);

app.use(compression());
app.use(
  cors({
    origin: process.env.ALLOWED_ORIGINS
      ? process.env.ALLOWED_ORIGINS.split(",")
      : "*",
    methods: ["GET", "HEAD"],
    maxAge: 86400,
  })
);
app.use(limiter);

// Monitoring endpoints
app.get("/metrics", async (req, res) => {
  try {
    res.set("Content-Type", register.contentType);
    res.end(await register.metrics());
  } catch (err) {
    logger.error("Error generating metrics", { error: err });
    res.status(500).end();
  }
});

// health endpoint
app.get("/health", async (req, res) => {
  try {
    await Promise.all([pool.query("SELECT 1"), redis.ping()]);
    res.json({ status: "healthy" });
  } catch (err) {
    logger.error("Health check failed", { error: err });
    res.status(503).json({ status: "unhealthy", error: err.message });
  }
});

serverAdapter.setBasePath("/admin/queues");
app.use("/admin/queues", serverAdapter.getRouter());

const getProjectBySubdomain = async (pool, subdomain) => {
  const cacheKey = `project:${subdomain}`;
  try {
    const cached = await redis.get(cacheKey);
    if (cached) {
      return { rows: [JSON.parse(cached)] };
    }

    const query = `
      SELECT id, subdomain, created_at
      FROM projects 
      WHERE subdomain = $1 
      LIMIT 1;
    `;
    const result = await pool.query(query, [subdomain]);

    if (result.rows.length > 0) {
      await redis.set(cacheKey, JSON.stringify(result.rows[0]), "EX", 300);
    }

    return result;
  } catch (err) {
    logger.error("Error getting project", { error: err, subdomain });
    throw err;
  }
};

const getLatestDeployment = async (pool, projectId) => {
  const cacheKey = `deployment:${projectId}`;
  try {
    const cached = await redis.get(cacheKey);
    if (cached) {
      return { rows: [JSON.parse(cached)] };
    }

    const query = `
      SELECT id, project_id, status, created_at
      FROM deployments 
      WHERE project_id = $1 AND status = 'READY' 
      ORDER BY created_at DESC 
      LIMIT 1;
    `;
    const result = await pool.query(query, [projectId]);

    if (result.rows.length > 0) {
      await redis.set(cacheKey, JSON.stringify(result.rows[0]), "EX", 60);
    }

    return result;
  } catch (err) {
    logger.error("Error getting deployment", { error: err, projectId });
    throw err;
  }
};

const proxy = httpProxy.createProxy();

proxy.on("proxyReq", (proxyReq, req, res) => {
  const url = req.url;
  if (url === "/") proxyReq.path += "index.html";
});

proxy.on("error", (err, req, res) => {
  logger.error("Proxy error details:", {
    error: err.message,
    code: err.code,
    hostname: req.hostname,
    path: req.path,
    target: `${S3_BASE_PATH}/${deploymentId}`,
  });

  if (!res.headersSent) {
    res.status(502).json({
      error: "Bad Gateway",
      details: process.env.NODE_ENV === "development" ? err.message : undefined,
    });
  }
});

const startServer = async () => {
  let pool;
  try {
    pool = await createPool();

    app.use(async (req, res) => {
      try {
        const hostname = req.hostname;
        const subdomain = hostname.split(".")[0];
        const startTime = Date.now();
        const end = httpRequestDuration.startTimer();

        const projectResult = await getProjectBySubdomain(pool, subdomain);
        if (projectResult.rows.length === 0) {
          logger.warn(`Subdomain not found`, { subdomain });
          end({ method: req.method, route: req.path, status_code: 404 });
          return res.status(404).json({ error: "Subdomain not found" });
        }

        const projectId = projectResult.rows[0].id;

        const deploymentResult = await getLatestDeployment(pool, projectId);
        if (deploymentResult.rows.length === 0) {
          logger.warn(`No deployment found`, { projectId });
          end({ method: req.method, route: req.path, status_code: 404 });
          return res
            .status(404)
            .json({ error: "No deployment found with status 'READY'" });
        }

        const deploymentId = deploymentResult.rows[0].id;

        const resolvesTo = `${S3_BASE_PATH}/${deploymentId}`;

        logger.info(`Proxying request`, {
          subdomain,
          projectId,
          deploymentId,
          resolvesTo,
          duration: Date.now() - startTime,
        });

        logger.info(`Proxying request to: ${resolvesTo}`);
        return proxy.web(req, res, {
          target: resolvesTo,
          changeOrigin: true,
          secure: false,
          headers: {
            Host: new URL(resolvesTo).host,
          },
        });
      } catch (error) {
        logger.error("Error handling request", { error: err, subdomain });
        end({ method: req.method, route: req.path, status_code: 500 });
        res.status(500).json({ error: "Internal Server Error" });
      }
    });

    const server = app.listen(process.env.PORT, () => {
      logger.info(`Server running on port ${process.env.PORT}`);
    });

    // Graceful shutdown
    const shutdown = async (signal) => {
      logger.info(`Received ${signal}. Starting graceful shutdown...`);

      server.close(async () => {
        logger.info("HTTP server closed");

        try {
          await Promise.all([
            pool.end(),
            redis.quit(),
            deploymentQueue.close(),
          ]);
          logger.info("All connections closed");
          process.exit(0);
        } catch (err) {
          logger.error("Error during shutdown", { error: err });
          process.exit(1);
        }
      });

      setTimeout(() => {
        logger.error("Forced shutdown after timeout");
        process.exit(1);
      }, 30000);
    };

    process.on("SIGTERM", () => shutdown("SIGTERM"));
    process.on("SIGINT", () => shutdown("SIGINT"));
  } catch (err) {
    logger.error("Failed to start server", { error: err });
    process.exit(1);
  }
};

startServer();
