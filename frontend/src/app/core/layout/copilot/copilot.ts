import { Component, ElementRef, inject, signal, ViewChild, AfterViewChecked } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CopilotService, CopilotMessage } from '../../services/copilot.service';
import { IntelligenceService } from '../../services/intelligence.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-copilot',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './copilot.html'
})
export class Copilot implements AfterViewChecked {
  @ViewChild('messagesContainer') messagesContainer!: ElementRef;

  copilot = inject(CopilotService);
  intelligence = inject(IntelligenceService);
  router = inject(Router);

  inputText = '';
  private shouldScrollToBottom = false;

  ngAfterViewChecked() {
    if (this.shouldScrollToBottom) {
      this.scrollToBottom();
      this.shouldScrollToBottom = false;
    }
  }

  scrollToBottom() {
    try {
      const el = this.messagesContainer?.nativeElement;
      if (el) el.scrollTop = el.scrollHeight;
    } catch (_) {}
  }

	sendMessage() {
		const text = this.inputText.trim();
		if (!text || this.copilot.isTyping()) return;

		this.inputText = '';
		this.shouldScrollToBottom = true;
		this.copilot.sendMessageToAI(text);
	}

  handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      this.sendMessage();
    }
  }

  formatContent(content: string): string {
    return content
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      .replace(/\n/g, '<br>');
  }
}
