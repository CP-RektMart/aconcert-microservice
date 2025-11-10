import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 50 },
    { duration: "1m", target: 50 },
    { duration: "30s", target: 200 },
    { duration: "1m", target: 200 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.02"],
    http_req_duration: ["p(95)<500"],
  },
};

const BASE_URL = "http://localhost:8000";
const ACCESS_TOKEN = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImQ3N2U5Njk0LTBhYmYtNDJmYi05MDlhLWFhM2MxMTc3ZDU3YyIsInVpZCI6ImExNDlmZmQzLTBmZjktNGEzNC1iNzBiLWNjMmZlODc0OTdhMSIsIk1hcENsYWltcyI6eyJleHAiOjE3NjMzODU5NzZ9fQ.dar83nCiSuCG-lJ42h4mvFrfvVmETJmu6sjBukAiNgE`; // pass via env var for security

export default function () {
  const headers = {
    Authorization: `Bearer ${ACCESS_TOKEN}`,
    "Content-Type": "application/json",
  };

  // 1️⃣ Get all locations
  const res = http.get(`${BASE_URL}/v1/locations`, { headers });
  check(res, { "GET /locations 200": (r) => r.status === 200 });

  const locations = res.json();
  if (locations && locations.length > 0) {
    // 2️⃣ Get specific location by ID
    const locationId =
      locations[Math.floor(Math.random() * locations.length)].id;
    const detailRes = http.get(`${BASE_URL}/locations/${locationId}`, {
      headers,
    });
    check(detailRes, { "GET /locations/:id 200": (r) => r.status === 200 });
  }

  sleep(1);
}
