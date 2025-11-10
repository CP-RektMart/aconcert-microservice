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
};

const BASE_URL = "http://localhost:8000";
const ACCESS_TOKEN = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImQ3N2U5Njk0LTBhYmYtNDJmYi05MDlhLWFhM2MxMTc3ZDU3YyIsInVpZCI6ImExNDlmZmQzLTBmZjktNGEzNC1iNzBiLWNjMmZlODc0OTdhMSIsIk1hcENsYWltcyI6eyJleHAiOjE3NjMzODU5NzZ9fQ.dar83nCiSuCG-lJ42h4mvFrfvVmETJmu6sjBukAiNgE`; // pass via env var for security

export default function () {
  const headers = {
    Authorization: `Bearer ${ACCESS_TOKEN}`,
    "Content-Type": "application/json",
  };

  // Example: reserve ticket
  const res = http.post(
    `${BASE_URL}/v1/reservations`,
    JSON.stringify({
      eventId: "b1a9d6b7-23f4-4b95-92f1-3a5e4f57e9aa",
      totalPrice: 300.0,
      seats: [
        {
          zoneNumber: 1,
          row: 10,
          column: 10,
        },
        {
          zoneNumber: 1,
          row: 10,
          column: 10,
        },
      ],
    }),
    { headers }
  );

  check(res, {
    "status is 200": (r) => r.status === 200,
  });

  sleep(1);
}
