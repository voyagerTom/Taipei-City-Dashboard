import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";
import viteCompression from "vite-plugin-compression";

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd()); // eslint-disable-line no-undef
	const isDockerCompose = process?.env.DOCKER_COMPOSE === "true"; // eslint-disable-line no-undef

	const twccProxy = {
		"/twcc-api": {
			target: env.VITE_TWCC_API_URL || "https://api-ams.twcc.ai/api",
			changeOrigin: true,
			secure: false,
			rewrite: (path) => path.replace(/^\/twcc-api/, ""),
			headers: {
				"X-API-KEY": env.VITE_TWCC_API_KEY || "",
			},
		},
	};

	const serverConfig = isDockerCompose
		? {
			host: "0.0.0.0",
			port: 80,
			proxy: {
				"/api/dev": {
					target: "http://dashboard-be:8080",
					changeOrigin: true,
					rewrite: (path) => path.replace("/dev", "/v1"),
				},
				...twccProxy,
			},
		}
		: {
			host: "0.0.0.0",
			port: 5173,
			proxy: {
				"/be-api": {
					target: "http://localhost:8088/api/v1",
					changeOrigin: true,
					rewrite: (path) => path.replace(/^\/be-api/, ""),
				},
				"/api": {
					target: "https://citydashboard.taipei/api/v1",
					changeOrigin: true,
					secure: false,
					rewrite: (path) => path.replace(/^\/api/, ""),
				},
				"/geo_server": {
					target: "https://citydashboard.taipei/geo_server/",
					changeOrigin: true,
					secure: false,
					rewrite: (path) => path.replace(/^\/geo_server/, ""),
				},
				...twccProxy,
			},
		};

	return {
		plugins: [vue(), viteCompression()],
		build: {
			rollupOptions: {
				output: {
					manualChunks(id) {
						if (id.includes("node_modules")) {
							return id
								.toString()
								.split("node_modules/")[1]
								.split("/")[0]
								.toString();
						}
					},
				},
			},
			chunkSizeWarningLimit: 1600,
		},
		base: "/",
		server: serverConfig,
	};
});
