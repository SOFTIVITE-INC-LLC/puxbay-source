import { inject, Injectable, signal } from '@angular/core';
import { ApiService } from './api.service';

export interface CopilotMessage {
	role: 'user' | 'assistant';
	content: string;
	timestamp: Date;
}

@Injectable({ providedIn: 'root' })
export class CopilotService {
	private api = inject(ApiService);

	isOpen = signal(false);
	messages = signal<CopilotMessage[]>([
		{
			role: 'assistant',
			content: "Hi! I'm **Puxbay Copilot**\n\nI can help you navigate the platform, understand your data, and give business insights. Try asking:\n- *\"How's my inventory looking?\"*\n- *\"Show me the top selling products\"*\n- *\"What should I reorder today?\"*",
			timestamp: new Date()
		}
	]);
	isTyping = signal(false);

	open() { this.isOpen.set(true); }
	close() { this.isOpen.set(false); }
	toggle() { this.isOpen.update(v => !v); }

	addMessage(role: 'user' | 'assistant', content: string) {
		this.messages.update(msgs => [...msgs, { role, content, timestamp: new Date() }]);
	}

	clearMessages() {
		this.messages.set([]);
	}

	sendMessageToAI(text: string) {
		this.addMessage('user', text);
		this.isTyping.set(true);

		// Prepare history payload for the backend (excluding the initial static message)
		const historyPayload = this.messages()
			.slice(1)
			.map(m => ({ role: m.role, content: m.content }));

		this.api.post<{ reply: string }>('/copilot/chat', { history: historyPayload }).subscribe({
			next: (res) => {
				this.addMessage('assistant', res.reply);
				this.isTyping.set(false);
			},
			error: (err) => {
				console.error('Copilot AI Error:', err);
				this.addMessage('assistant', "I'm sorry, I encountered an error connecting to my AI brain. Please check the backend configuration or your internet connection.");
				this.isTyping.set(false);
			}
		});
	}
}
