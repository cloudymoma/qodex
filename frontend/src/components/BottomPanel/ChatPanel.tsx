import { useState } from 'react';
import { Send } from 'lucide-react';

interface Message {
  role: 'user' | 'assistant';
  content: string;
}

export function ChatPanel() {
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState<Message[]>([
    {
      role: 'assistant',
      content: 'Hello! I can help you understand this codebase. Ask me anything about the code structure, dependencies, or architecture. (Ollama integration coming in Phase 2)',
    },
  ]);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;

    const userMsg: Message = { role: 'user', content: input.trim() };
    const botMsg: Message = {
      role: 'assistant',
      content: 'Chat with Ollama is not yet connected. This is a placeholder for Phase 2 integration.',
    };

    setMessages((prev) => [...prev, userMsg, botMsg]);
    setInput('');
  };

  return (
    <div className="flex flex-col h-full">
      {/* Messages */}
      <div className="flex-1 overflow-auto px-4 py-2 space-y-2">
        {messages.map((msg, i) => (
          <div
            key={i}
            className={`text-sm px-3 py-1.5 rounded ${
              msg.role === 'user'
                ? 'bg-accent-primary/20 text-dark-text ml-8'
                : 'bg-dark-bg-tertiary text-dark-text-secondary mr-8'
            }`}
          >
            {msg.content}
          </div>
        ))}
      </div>

      {/* Input */}
      <form onSubmit={handleSend} className="flex items-center gap-2 px-4 py-2 border-t border-dark-border">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Ask about the codebase..."
          className="flex-1 bg-dark-bg-tertiary text-dark-text px-3 py-1 rounded border border-dark-border focus:border-accent-primary focus:outline-none text-sm"
        />
        <button
          type="submit"
          className="p-1.5 bg-accent-primary text-white rounded hover:bg-accent-primary/80"
        >
          <Send size={14} />
        </button>
      </form>
    </div>
  );
}
