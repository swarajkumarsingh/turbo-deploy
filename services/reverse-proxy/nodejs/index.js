require("dotenv").config();
const { Pool } = require("pg");
const express = require("express");
const httpProxy = require("http-proxy");

const PORT = 8000;
const app = express();
const BASE_PATH = process.env.BASE_PATH;

const pool = new Pool({
  host: process.env.DB_HOST,
  port: process.env.DB_HOST,
  user: process.env.DB_USER,
  database: process.env.DB_DATABASE,
  password: process.env.DB_PASSWORD,
  ssl: { rejectUnauthorized: false },
});

const proxy = httpProxy.createProxy();

app.use(async (req, res) => {
  try {
    const hostname = req.hostname;
    const subdomain = hostname.split(".")[0];

    const projectQuery = `
      SELECT id 
      FROM projects 
      WHERE subdomain = $1 
      LIMIT 1;
    `;
    const projectResult = await pool.query(projectQuery, [subdomain]);

    if (projectResult.rows.length === 0) {
      return res.status(404).send("Subdomain not found");
    }
    const projectId = projectResult.rows[0].id;

    const deploymentQuery = `
      SELECT id 
      FROM deployments 
      WHERE project_id = $1 AND status = 'READY' 
      ORDER BY created_at DESC 
      LIMIT 1;
    `;
    const deploymentResult = await pool.query(deploymentQuery, [projectId]);

    if (deploymentResult.rows.length === 0) {
      return res.status(404).send("No deployment found with status 'READY'");
    }
    const deploymentId = deploymentResult.rows[0].id;

    const resolvesTo = `${BASE_PATH}/${deploymentId}`;
    return proxy.web(req, res, { target: resolvesTo, changeOrigin: true });
  } catch (err) {
    console.error("Error handling request:", err);
    res.status(500).send("Internal Server Error");
  }
});

proxy.on("proxyReq", (proxyReq, req, res) => {
  const url = req.url;
  if (url === "/") proxyReq.path += "index.html";
});

app.listen(PORT, () => console.log(`Reverse Proxy Running on Port ${PORT}`));
