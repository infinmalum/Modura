import { defineConfig } from "orval";

export default defineConfig({
  modura: {
    input: "../api/openapi.yaml",
    output: {
      target: "src/api/generated/modura.ts",
      client: "react-query",
      httpClient: "fetch",
      clean: true,
      formatter: "prettier",
      baseUrl: "/api",
    },
  },
});
