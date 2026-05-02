import { ref, computed, watch } from "vue";
import { defineStore } from "pinia";
import { useContentStore } from "./contentStore";
import axios from "axios";

const beApi = axios.create({ baseURL: "/be-api" });

const TWCC_MODEL = import.meta.env.VITE_TWCC_MODEL || "llama3.3-ffm-70b-16k-chat";

const CAT_SYSTEM_PROMPT = `你是「小駭」，雙北城市儀表板的 AI 狐狸助理。

## 你的性格
- 自稱「小駭」，稱呼使用者「你」或「大家」
- 語氣親切活潑，像鄰居朋友在聊天，偶爾加「～」「！」
- 偶爾用一個顏文字：(๑•̀ㅂ•́)و✧、꒰ᐢ⸝⸝•༝•⸝⸝ᐢ꒱、( •̀ᴗ•́ )
- 說話口語化，用「白話」解釋數據，避免專業術語
- 把數字轉換成一般人能感受的說法（例：「大概每 5 個人就有 1 個超過 65 歲」而非「老年人口佔比 20%」）

## 核心原則
- 你的任務是讓「完全不懂數據的人」也能秒懂重點
- 嚴禁使用：扶養比、指數、佔比、同比、環比、趨勢分析 等專業詞彙
- 用生活化的比喻和具體場景來傳達資訊
- 所有描述必須基於提供的真實數據，嚴禁編造

## 回應格式
你必須回傳純 JSON（不要加 markdown code block），格式如下：
{
  "bubble": "一句話講完重點，像朋友丟訊息給你（15-25字）",
  "summary": {
    "overview": "用 2-3 句白話告訴大家「現在是什麼狀況」，像在跟朋友解釋",
    "warnings": "有什麼要小心的？用生活情境說明（2-3句）",
    "suggestions": "小駭給大家的實用建議，具體可以怎麼做（2-3句）"
  },
  "quick_replies": [
    {"label": "按鈕文字（口語化）", "prompt": "對應的延伸提問"}
  ]
}

## 完整回應範例
{
  "bubble": "板橋跟中和最近人變多了！通勤要注意～",
  "summary": {
    "overview": "最近雙北的人口有在慢慢增加，尤其板橋和中和特別明顯。感覺大家都往新北市區搬，捷運站附近的人潮越來越多了～",
    "warnings": "上下班時間搭藍線的話，板橋站和新埔站會比較擠。如果可以的話，避開早上 8 點到 9 點那段時間會舒服很多！",
    "suggestions": "小駭建議通勤族可以試試看提早 15 分鐘出門，或是改搭黃線轉車，人會少很多！假日要去板橋的話，下午 2 點前到比較不會塞～"
  },
  "quick_replies": [
    {"label": "哪個時段人最少？", "prompt": "哪個時段搭捷運人最少最舒適？"},
    {"label": "跟上個月比怎樣？", "prompt": "跟上個月的數據比起來有什麼變化？"},
    {"label": "有什麼替代路線？", "prompt": "有沒有比較不擠的替代通勤路線推薦？"}
  ]
}

## 規則
- bubble 要像 LINE 訊息一樣自然，不超過 25 字
- summary 三段都要有實質內容，但用大家聽得懂的話說
- quick_replies 提供 2-3 個按鈕，文字要像你會想點的東西
- 嚴禁回傳 markdown code block，只回傳純 JSON
- 所有內容都要用繁體中文`;

export const useCatStore = defineStore("cat", () => {
	const contentStore = useContentStore();

	// Cat visual state
	const catState = ref("idle");
	const isPanelOpen = ref(false);

	// LLM data
	const bubbleMessage = ref("");
	const showBubble = ref(false);
	const summary = ref(null);
	const quickReplies = ref([]);
	const isLoading = ref(false);
	const lastError = ref(null);

	// Chat history
	const chatHistory = ref([]);

	let bubbleTimer = null;

	// Track last loaded dashboard
	let lastDashboardKey = null;
	let backendAvailable = null; // null = unknown, true/false after first attempt

	// Session-level cache: avoids API calls when switching back to an already-loaded dashboard
	const sessionCache = new Map();

	const currentDashboardKey = computed(() => {
		const { index, city } = contentStore.currentDashboard || {};
		return index && city ? `${index}:${city}` : null;
	});

	// --- TWCC direct fallback ---

	function collectDashboardData() {
		const { name, index, city, components } = contentStore.currentDashboard;
		if (!components || components.length === 0) return null;

		const summaries = [];
		for (const comp of components) {
			if (!comp.chart_data || comp.chart_data.length === 0) continue;
			const data = comp.chart_data;
			const sample = Array.isArray(data)
				? data.slice(0, 8).map((item) => {
					const s = {};
					for (const [k, v] of Object.entries(item)) {
						if (k === "data" && typeof v === "string") continue;
						if (typeof v === "number" || typeof v === "string") s[k] = v;
					}
					return s;
				})
				: [];
			summaries.push({ name: comp.name, source: comp.source, data_sample: sample });
		}

		return {
			dashboard_name: name || index,
			city: city || "taipei",
			components: summaries.slice(0, 12),
			timestamp: new Date().toLocaleString("zh-TW", { timeZone: "Asia/Taipei" }),
		};
	}

	async function callTWCCDirect(userMessage) {
		const response = await axios.post("/twcc-api/models/conversation", {
			model: TWCC_MODEL,
			messages: [
				{ role: "system", content: CAT_SYSTEM_PROMPT },
				{ role: "user", content: userMessage },
			],
			parameters: { max_new_tokens: 600, temperature: 0.8, top_p: 0.9 },
		});

		let content = "";
		if (response.data?.generated_text) content = response.data.generated_text;
		else if (response.data?.choices?.[0]?.message?.content) content = response.data.choices[0].message.content;
		return parseResponse(content);
	}

	function parseResponse(content) {
		if (!content) return null;
		content = content.trim().replace(/^```json\s*/i, "").replace(/```\s*$/, "").trim();

		try {
			const parsed = JSON.parse(content);
			if (parsed.bubble) return parsed;
		} catch { /* fallback */ }

		const match = content.match(/\{[\s\S]*"bubble"[\s\S]*\}/);
		if (match) {
			try { const p = JSON.parse(match[0]); if (p.bubble) return p; } catch { /* fallback */ }
		}

		return {
			bubble: content.slice(0, 60),
			summary: { overview: content, warnings: "", suggestions: "" },
			quick_replies: [],
		};
	}

	// --- Backend API (preferred) ---

	async function tryBackendSummary() {
		const { index, city } = contentStore.currentDashboard || {};
		if (!index) return null;

		const response = await beApi.get(`/ai/summary/${index}`, { params: { city: city || "taipei" }, timeout: 120000 });
		return response.data?.data || null;
	}

	async function tryBackendChat(question) {
		const { index, city } = contentStore.currentDashboard || {};
		if (!index) return null;

		const payload = {
			dashboard_index: index,
			city: city || "taipei",
			question,
		};

		if (summary.value) {
			payload.context = JSON.stringify(summary.value);
		}

		const response = await beApi.post("/ai/chat/public", payload, { timeout: 120000 });
		return response.data?.data || null;
	}

	// --- Main logic ---

	function applyData(data) {
		bubbleMessage.value = data.bubble || "";
		summary.value = data.summary || null;
		quickReplies.value = data.quick_replies || [];
		showBubbleTemporarily();
		catState.value = "talking";
		setTimeout(() => { if (catState.value === "talking") catState.value = "idle"; }, 5000);
	}

	async function fetchSummary() {
		const { index, city } = contentStore.currentDashboard || {};
		if (!index) return;

		const key = `${index}:${city || "taipei"}`;
		if (key === lastDashboardKey && summary.value) return;

		// Session cache hit: restore instantly without loading animation
		if (sessionCache.has(key)) {
			const cached = sessionCache.get(key);
			lastDashboardKey = key;
			applyData(cached);
			return;
		}

		isLoading.value = true;
		catState.value = "thinking";
		lastError.value = null;

		try {
			let data = null;

			// Try backend first (if not known to be unavailable)
			if (backendAvailable !== false) {
				try {
					data = await tryBackendSummary();
					backendAvailable = true;
				} catch {
					backendAvailable = false;
				}
			}

			// Fallback to TWCC direct
			if (!data) {
				const dashData = collectDashboardData();
				if (dashData) {
					const prompt = `以下是「${dashData.dashboard_name}」儀表板的即時數據，請根據這些真實數據給出評論：\n\n${JSON.stringify(dashData)}`;
					data = await callTWCCDirect(prompt);
				}
			}

			if (data) {
				lastDashboardKey = key;
				sessionCache.set(key, data);
				applyData(data);
			}
		} catch (err) {
			console.error("CatStore fetchSummary error:", err);
			lastError.value = err.message || "連線失敗";
			catState.value = "idle";
		} finally {
			isLoading.value = false;
		}
	}

	async function askQuestion(question) {
		if (!question?.trim()) return;

		chatHistory.value.push({ role: "user", content: question });
		isLoading.value = true;
		catState.value = "thinking";
		lastError.value = null;

		try {
			let data = null;

			if (backendAvailable) {
				try { data = await tryBackendChat(question); } catch { /* fallback */ }
			}

			if (!data) {
				const dashData = collectDashboardData();
				const contextStr = dashData ? JSON.stringify(dashData) : "（無數據）";
				const prompt = `基於儀表板數據：${contextStr}\n\n使用者追問：${question}`;
				data = await callTWCCDirect(prompt);
			}

			if (data) {
				const answer = data.bubble || data.summary?.overview || "本喵想不出來喵...";
				chatHistory.value.push({ role: "cat", content: answer, full: data });
				bubbleMessage.value = data.bubble || "";
				summary.value = data.summary || summary.value;
				quickReplies.value = data.quick_replies || [];
				showBubbleTemporarily();
			}
		} catch (err) {
			console.error("CatStore askQuestion error:", err);
			chatHistory.value.push({ role: "cat", content: "嗚...連線出了點問題" });
			lastError.value = err.message || "連線失敗";
		} finally {
			isLoading.value = false;
			catState.value = "idle";
		}
	}

	async function handleQuickReply(prompt) {
		isPanelOpen.value = true;
		await askQuestion(prompt);
	}

	// --- UI helpers ---

	function showBubbleTemporarily() {
		showBubble.value = true;
		clearTimeout(bubbleTimer);
		bubbleTimer = setTimeout(() => { showBubble.value = false; }, 8000);
	}

	function togglePanel() {
		isPanelOpen.value = !isPanelOpen.value;
	}

	function initialize() {
		watch(currentDashboardKey, (newKey, oldKey) => {
			if (newKey && newKey !== oldKey) {
				chatHistory.value = [];
				lastDashboardKey = null;

				if (sessionCache.has(newKey)) {
					// Cached: restore instantly, no delay
					fetchSummary();
				} else {
					// Not cached: show loading after short delay
					summary.value = null;
					setTimeout(() => fetchSummary(), 2500);
				}
			}
		}, { immediate: true });
	}

	function cleanup() {
		clearTimeout(bubbleTimer);
		bubbleTimer = null;
	}

	return {
		catState, isPanelOpen,
		bubbleMessage, showBubble, summary, quickReplies, isLoading, lastError,
		chatHistory,
		initialize, cleanup, fetchSummary, askQuestion, handleQuickReply, togglePanel, showBubbleTemporarily,
	};
});
