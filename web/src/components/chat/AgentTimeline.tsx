import { agentLabel } from '../../lib/agents';

interface AgentTimelineProps {
  agents: string[];
  /** 1-based current step from task; 0 = none running yet */
  currentStep?: number;
  /** final status if known */
  status?: 'pending' | 'running' | 'completed' | 'failed' | 'awaiting' | 'cancelled' | string;
}

export function AgentTimeline({ agents, currentStep = 0, status }: AgentTimelineProps) {
  if (!agents?.length) return null;

  return (
    <div className="mt-3 rounded-2xl border border-ink/8 bg-canvas/80 p-3">
      <p className="text-[11px] font-semibold uppercase tracking-wider text-ink/40 mb-2.5">
        Langkah pengerjaan
      </p>
      <ol className="flex flex-wrap items-center gap-2">
        {agents.map((agent, idx) => {
          const stepNum = idx + 1;
          let tone = 'bg-white border-ink/10 text-ink/40';
          let mark = '○';

          if (status === 'awaiting' || status === 'cancelled' || status === 'pending') {
            tone = 'bg-white border-ink/10 text-ink/45';
            mark = '○';
          } else if (status === 'completed') {
            tone = 'bg-leaf/10 border-leaf/40 text-leaf';
            mark = '✓';
          } else if (status === 'failed' && currentStep === stepNum) {
            tone = 'bg-rose-50 border-rose-300 text-rose-700';
            mark = '!';
          } else if (status === 'failed' && currentStep > stepNum) {
            tone = 'bg-leaf/10 border-leaf/40 text-leaf';
            mark = '✓';
          } else if (status === 'failed' && currentStep < stepNum) {
            tone = 'bg-white border-ink/10 text-ink/30';
            mark = '○';
          } else if (currentStep > stepNum) {
            tone = 'bg-leaf/10 border-leaf/40 text-leaf';
            mark = '✓';
          } else if (currentStep === stepNum && status === 'running') {
            tone = 'bg-banana/30 border-banana text-ink font-semibold';
            mark = '…';
          }

          return (
            <li key={`${agent}-${idx}`} className="flex items-center gap-2">
              <span className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] ${tone}`}>
                <span aria-hidden>{mark}</span>
                {agentLabel(agent)}
              </span>
              {idx < agents.length - 1 && (
                <span className="text-ink/20 text-xs" aria-hidden>
                  →
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </div>
  );
}
