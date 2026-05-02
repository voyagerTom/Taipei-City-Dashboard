<script setup>
import { onMounted, onBeforeUnmount, ref, nextTick } from "vue";
import { useCatStore } from "../../store/catStore";
import catIcon from "../../assets/images/cat-assistant.svg";

const catStore = useCatStore();
const chatInput = ref("");
const chatBodyRef = ref(null);

onMounted(() => {
	catStore.initialize();
});

onBeforeUnmount(() => {
	catStore.cleanup();
});

function onCatClick() {
	catStore.togglePanel();
}

function closePanel() {
	catStore.isPanelOpen = false;
}

async function sendMessage() {
	const msg = chatInput.value.trim();
	if (!msg || catStore.isLoading) return;
	chatInput.value = "";
	await catStore.askQuestion(msg);
	await nextTick();
	scrollChatToBottom();
}

async function onQuickReply(prompt) {
	await catStore.handleQuickReply(prompt);
	await nextTick();
	scrollChatToBottom();
}

function scrollChatToBottom() {
	if (chatBodyRef.value) {
		chatBodyRef.value.scrollTop = chatBodyRef.value.scrollHeight;
	}
}

function handleKeydown(e) {
	if (e.key === "Enter" && !e.shiftKey) {
		e.preventDefault();
		sendMessage();
	}
}
</script>

<template>
  <div class="cat-assistant">
    <!-- Summary Panel -->
    <Transition name="panel-slide">
      <div
        v-if="catStore.isPanelOpen"
        class="cat-panel"
      >
        <div class="cat-panel-header">
          <div class="cat-panel-header-left">
            <img
              :src="catIcon"
              alt="小駭"
              class="panel-header-icon"
            >
            <span class="cat-panel-title">小駭報你知</span>
          </div>
          <button
            class="cat-panel-close"
            @click="closePanel"
          >
            ✕
          </button>
        </div>

        <div
          ref="chatBodyRef"
          class="cat-panel-body"
        >
          <!-- Summary Section -->
          <template v-if="catStore.chatHistory.length === 0">
            <template v-if="catStore.isLoading && !catStore.summary">
              <div class="cat-panel-loading">
                <div class="loading-dots">
                  <span /><span /><span />
                </div>
                <p>小駭正在幫你看數據～</p>
              </div>
            </template>

            <template v-else-if="catStore.summary">
              <div class="cat-panel-section">
                <div class="section-icon">
                  📊
                </div>
                <div class="section-content">
                  <h4>整體狀況</h4>
                  <p>{{ catStore.summary.overview }}</p>
                </div>
              </div>
              <div class="cat-panel-section warning">
                <div class="section-icon">
                  ⚠️
                </div>
                <div class="section-content">
                  <h4>需要注意</h4>
                  <p>{{ catStore.summary.warnings }}</p>
                </div>
              </div>
              <div class="cat-panel-section suggestion">
                <div class="section-icon">
                  💡
                </div>
                <div class="section-content">
                  <h4>小駭建議</h4>
                  <p>{{ catStore.summary.suggestions }}</p>
                </div>
              </div>
            </template>

            <template v-else-if="catStore.lastError">
              <div class="cat-panel-error">
                <p>嗚...連線出了點問題 ꒰ᐢ⸝⸝•༝•⸝⸝ᐢ꒱</p>
                <button
                  class="retry-btn"
                  @click="catStore.fetchSummary()"
                >
                  再試一次
                </button>
              </div>
            </template>

            <template v-else>
              <div class="cat-panel-empty">
                <p>小駭還在等數據載入～</p>
              </div>
            </template>
          </template>

          <!-- Chat History -->
          <template v-else>
            <div
              v-for="(msg, i) in catStore.chatHistory"
              :key="i"
              class="chat-message"
              :class="msg.role"
            >
              <img
                v-if="msg.role === 'cat'"
                :src="catIcon"
                alt="小駭"
                class="chat-avatar-img"
              >
              <div class="chat-bubble-msg">
                {{ msg.content }}
              </div>
            </div>
            <div
              v-if="catStore.isLoading"
              class="chat-message cat"
            >
              <img
                :src="catIcon"
                alt="小駭"
                class="chat-avatar-img"
              >
              <div class="chat-bubble-msg thinking-msg">
                小駭想想
                <span class="dot">.</span><span class="dot">.</span><span class="dot">.</span>
              </div>
            </div>
          </template>
        </div>

        <!-- Quick Replies -->
        <div
          v-if="catStore.quickReplies?.length > 0 && !catStore.isLoading"
          class="cat-panel-quick"
        >
          <button
            v-for="(reply, i) in catStore.quickReplies"
            :key="i"
            class="quick-reply-btn"
            :disabled="catStore.isLoading"
            @click="onQuickReply(reply.prompt)"
          >
            {{ reply.label }}
          </button>
        </div>

        <!-- Input Area -->
        <div class="cat-panel-input">
          <input
            v-model="chatInput"
            type="text"
            placeholder="問小駭任何問題..."
            :disabled="catStore.isLoading"
            @keydown="handleKeydown"
          >
          <button
            class="send-btn"
            :disabled="!chatInput.trim() || catStore.isLoading"
            @click="sendMessage"
          >
            <span class="send-icon">➤</span>
          </button>
        </div>
      </div>
    </Transition>

    <!-- Speech Bubble (appears above the cat icon) -->
    <Transition name="bubble-fade">
      <div
        v-if="catStore.showBubble && catStore.bubbleMessage && !catStore.isPanelOpen"
        class="cat-bubble"
      >
        <div class="bubble-content">
          {{ catStore.bubbleMessage }}
        </div>
        <div class="bubble-tail" />
      </div>
    </Transition>

    <!-- Thinking Bubble -->
    <Transition name="bubble-fade">
      <div
        v-if="catStore.isLoading && !catStore.showBubble && !catStore.isPanelOpen"
        class="cat-bubble thinking"
      >
        <div class="bubble-content thinking-msg">
          小駭想想
          <span class="dot">.</span><span class="dot">.</span><span class="dot">.</span>
        </div>
        <div class="bubble-tail" />
      </div>
    </Transition>

    <!-- Cat Icon Button -->
    <div
      class="cat-character"
      :class="{ 'cat-active': catStore.isPanelOpen }"
      @click="onCatClick"
    >
      <img
        :src="catIcon"
        alt="小駭"
        class="cat-icon"
      >
      <div
        v-if="catStore.isLoading"
        class="cat-pulse-ring"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
$bg-primary: #090909;
$bg-component: #282a2c;
$border-color: #494b4e;
$text-body: #888787;
$text-title: #ffffff;
$accent: #5a9cf8;
$accent-dim: rgba(90, 156, 248, 0.15);
$accent-border: rgba(90, 156, 248, 0.3);

.cat-assistant {
	display: flex;
	flex-direction: column;
	align-items: flex-end;
	position: relative;
}

// === Cat Icon ===
.cat-character {
	width: 70px;
	height: 70px;
	border-radius: 50%;
	cursor: pointer;
	display: flex;
	align-items: center;
	justify-content: center;
	background: $bg-component;
	border: 2px solid $border-color;
	transition: all 0.25s ease;
	position: relative;
	flex-shrink: 0;

	&:hover {
		border-color: $accent;
		transform: scale(1.08);
		box-shadow: 0 0 16px rgba(90, 156, 248, 0.25);
	}

	&.cat-active {
		border-color: $accent;
		box-shadow: 0 0 12px rgba(90, 156, 248, 0.2);
	}
}

.cat-icon {
	width: 48px;
	height: 48px;
	border-radius: 50%;
}

.cat-pulse-ring {
	position: absolute;
	inset: -4px;
	border-radius: 50%;
	border: 2px solid $accent;
	animation: pulse-ring 1.5s ease-out infinite;
}
@keyframes pulse-ring {
	0% { transform: scale(1); opacity: 0.6; }
	100% { transform: scale(1.4); opacity: 0; }
}

// === Speech Bubble ===
.cat-bubble {
	position: absolute;
	bottom: 82px;
	right: 0;
	max-width: 260px;
	z-index: 10;

	.bubble-content {
		background: $bg-component;
		color: $text-title;
		padding: 10px 14px;
		border-radius: 12px;
		font-size: 13px;
		line-height: 1.5;
		border: 1px solid $border-color;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
	}

	.bubble-tail {
		width: 10px;
		height: 10px;
		background: $bg-component;
		transform: rotate(45deg);
		position: absolute;
		bottom: -5px;
		right: 20px;
		border-right: 1px solid $border-color;
		border-bottom: 1px solid $border-color;
	}

	&.thinking .bubble-content {
		background: rgba(40, 42, 44, 0.95);
		color: $text-body;
	}
}

.thinking-msg .dot {
	animation: dot-blink 1.4s ease-in-out infinite;
	&:nth-child(2) { animation-delay: 0.2s; }
	&:nth-child(3) { animation-delay: 0.4s; }
}
@keyframes dot-blink {
	0%, 20% { opacity: 0; }
	50% { opacity: 1; }
	100% { opacity: 0; }
}

.bubble-fade-enter-active { animation: bubble-in 0.25s ease-out; }
.bubble-fade-leave-active { animation: bubble-out 0.2s ease-in; }
@keyframes bubble-in {
	from { opacity: 0; transform: translateY(6px) scale(0.9); }
	to { opacity: 1; transform: translateY(0) scale(1); }
}
@keyframes bubble-out {
	from { opacity: 1; }
	to { opacity: 0; transform: translateY(-6px) scale(0.9); }
}

// === Summary Panel ===
.cat-panel {
	position: fixed;
	bottom: 110px;
	right: 24px;
	width: 380px;
	max-height: 70vh;
	background: $bg-primary;
	border-radius: 12px;
	border: 1px solid $border-color;
	box-shadow: 0 8px 40px rgba(0, 0, 0, 0.5);
	display: flex;
	flex-direction: column;
	z-index: 1002;
	overflow: hidden;
}

.cat-panel-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 12px 16px;
	border-bottom: 1px solid $border-color;
	background: $bg-component;
	flex-shrink: 0;
}

.cat-panel-header-left {
	display: flex;
	align-items: center;
	gap: 8px;
}

.panel-header-icon {
	width: 24px;
	height: 24px;
	border-radius: 50%;
}

.cat-panel-title {
	font-size: 14px;
	font-weight: 600;
	color: $text-title;
}

.cat-panel-close {
	width: 26px;
	height: 26px;
	border-radius: 6px;
	display: flex;
	align-items: center;
	justify-content: center;
	color: $text-body;
	font-size: 13px;
	transition: all 0.2s;

	&:hover {
		background: rgba(255, 255, 255, 0.08);
		color: $text-title;
	}
}

.cat-panel-body {
	padding: 14px 16px;
	overflow-y: auto;
	flex: 1;
	min-height: 0;

	&::-webkit-scrollbar { width: 4px; }
	&::-webkit-scrollbar-track { background: transparent; }
	&::-webkit-scrollbar-thumb {
		background: $border-color;
		border-radius: 2px;
	}
}

.cat-panel-section {
	display: flex;
	gap: 10px;
	padding: 10px 0;

	&:not(:last-child) {
		border-bottom: 1px solid rgba(73, 75, 78, 0.5);
	}

	.section-icon {
		font-size: 18px;
		flex-shrink: 0;
		padding-top: 1px;
	}

	.section-content {
		flex: 1;

		h4 {
			font-size: 13px;
			font-weight: 600;
			color: $accent;
			margin-bottom: 4px;
		}

		p {
			font-size: 13px;
			line-height: 1.7;
			color: rgba(255, 255, 255, 0.78);
		}
	}
}

// === Chat Messages ===
.chat-message {
	display: flex;
	gap: 8px;
	margin-bottom: 10px;
	align-items: flex-start;

	&.user {
		justify-content: flex-end;
	}

	&.user .chat-bubble-msg {
		background: $accent-dim;
		color: rgba(255, 255, 255, 0.9);
		border: 1px solid $accent-border;
		border-radius: 12px 12px 4px 12px;
	}

	&.cat .chat-bubble-msg {
		background: $bg-component;
		color: rgba(255, 255, 255, 0.85);
		border: 1px solid $border-color;
		border-radius: 12px 12px 12px 4px;
	}
}

.chat-avatar-img {
	width: 26px;
	height: 26px;
	border-radius: 50%;
	flex-shrink: 0;
	margin-top: 2px;
}

.chat-bubble-msg {
	padding: 8px 12px;
	font-size: 13px;
	line-height: 1.6;
	max-width: 82%;
}

// === Quick Replies ===
.cat-panel-quick {
	padding: 8px 16px;
	display: flex;
	flex-wrap: wrap;
	gap: 6px;
	border-top: 1px solid $border-color;
	flex-shrink: 0;
}

.quick-reply-btn {
	padding: 5px 12px;
	border-radius: 14px;
	font-size: 12px;
	color: $accent;
	background: $accent-dim;
	border: 1px solid $accent-border;
	transition: all 0.2s;
	cursor: pointer;

	&:hover {
		background: rgba(90, 156, 248, 0.25);
		border-color: $accent;
	}
}

// === Input Area ===
.cat-panel-input {
	display: flex;
	gap: 8px;
	padding: 10px 12px;
	border-top: 1px solid $border-color;
	flex-shrink: 0;
	background: $bg-component;

	input {
		flex: 1;
		padding: 8px 14px;
		border-radius: 18px;
		font-size: 13px;
		background: rgba(255, 255, 255, 0.06);
		color: $text-title;
		border: 1px solid $border-color;
		outline: none;

		&::placeholder { color: $text-body; }

		&:focus {
			border-color: $accent;
			background: rgba(255, 255, 255, 0.08);
		}

		&:disabled { opacity: 0.45; }
	}
}

.send-btn {
	width: 34px;
	height: 34px;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	background: $accent;
	color: white;
	font-size: 13px;
	transition: all 0.2s;
	flex-shrink: 0;

	&:hover:not(:disabled) {
		background: darken($accent, 8%);
		box-shadow: 0 2px 8px rgba(90, 156, 248, 0.3);
	}

	&:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}
}

.send-icon {
	display: inline-block;
}

// === States ===
.cat-panel-loading {
	display: flex;
	flex-direction: column;
	align-items: center;
	padding: 28px 0;
	gap: 10px;

	p {
		color: $text-body;
		font-size: 13px;
	}
}

.loading-dots {
	display: flex;
	gap: 5px;

	span {
		width: 7px;
		height: 7px;
		background: $accent;
		border-radius: 50%;
		animation: loading-bounce 1.2s ease-in-out infinite;

		&:nth-child(2) { animation-delay: 0.15s; }
		&:nth-child(3) { animation-delay: 0.3s; }
	}
}
@keyframes loading-bounce {
	0%, 80%, 100% { transform: scale(0.6); opacity: 0.35; }
	40% { transform: scale(1); opacity: 1; }
}

.cat-panel-error {
	text-align: center;
	padding: 24px 0;

	p {
		color: rgba(255, 255, 255, 0.65);
		font-size: 13px;
		margin-bottom: 12px;
	}
}

.retry-btn {
	padding: 7px 18px;
	border-radius: 18px;
	background: $accent-dim;
	color: $accent;
	font-size: 12px;
	border: 1px solid $accent-border;
	cursor: pointer;
	transition: all 0.2s;

	&:hover {
		background: rgba(90, 156, 248, 0.25);
	}
}

.cat-panel-empty {
	text-align: center;
	padding: 24px 0;

	p {
		color: $text-body;
		font-size: 13px;
		line-height: 1.8;
	}
}

// === Panel Transition ===
.panel-slide-enter-active {
	animation: panel-in 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.panel-slide-leave-active {
	animation: panel-out 0.2s ease-in;
}
@keyframes panel-in {
	from { opacity: 0; transform: translateY(12px) scale(0.96); }
	to { opacity: 1; transform: translateY(0) scale(1); }
}
@keyframes panel-out {
	from { opacity: 1; }
	to { opacity: 0; transform: translateY(12px) scale(0.96); }
}

@media (max-width: 750px) {
	.cat-assistant { display: none; }
}
@media (max-width: 960px) {
	.cat-panel { width: 320px; }
}
</style>
