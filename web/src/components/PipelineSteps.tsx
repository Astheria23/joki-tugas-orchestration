import { CheckCircle2, Loader2, SkipForward, AlertTriangle, Circle } from 'lucide-react';
import { agentLabel } from '../lib/agents';
import type { Task } from '../lib/types';

export function PipelineSteps({ task }: { task: Task }) {
  if (!task.pipeline?.length) {
    return (
      <p className="text-sm text-ink/50">
        {task.status === 'pending' || task.status === 'running'
          ? 'Menyusun urutan agen…'
          : 'Belum ada pipeline.'}
      </p>
    );
  }

  return (
    <ol className="flex flex-wrap items-center gap-2">
      {task.pipeline.map((step, idx) => {
        const history = task.history?.find((h) => h.agentKey?.toLowerCase() === step.toLowerCase());
        const isCurrent = task.currentStep === idx + 1 && (task.status === 'running' || task.status === 'pending');
        const isDone =
          history?.status === 'success' ||
          (task.currentStep > idx + 1 && history?.status !== 'failed' && history?.status !== 'skipped') ||
          (task.status === 'completed' && (!history || history.status === 'success'));
        const isFailed = history?.status === 'failed' || (isCurrent && task.status === 'failed');
        const isSkipped = history?.status === 'skipped';

        let tone =
          'bg-white border-ink/10 text-ink/40';
        if (isFailed) tone = 'bg-rose-50 border-rose-400 text-rose-700';
        else if (isSkipped) tone = 'bg-amber-50 border-amber-400 text-amber-800 line-through';
        else if (isDone) tone = 'bg-leaf/10 border-leaf/40 text-leaf';
        else if (isCurrent) tone = 'bg-banana/30 border-banana text-ink font-semibold ring-2 ring-banana/40';

        return (
          <li key={`${step}-${idx}`} className="flex items-center gap-2">
            <span className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full border text-xs ${tone}`}>
              {isDone && !isSkipped && <CheckCircle2 className="h-3.5 w-3.5" />}
              {isCurrent && !isFailed && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
              {isSkipped && <SkipForward className="h-3.5 w-3.5" />}
              {isFailed && <AlertTriangle className="h-3.5 w-3.5" />}
              {!isDone && !isCurrent && !isSkipped && !isFailed && <Circle className="h-3 w-3 opacity-40" />}
              <span>{agentLabel(step)}</span>
            </span>
            {idx < task.pipeline.length - 1 && (
              <span className="text-ink/20 text-sm" aria-hidden>
                →
              </span>
            )}
          </li>
        );
      })}
    </ol>
  );
}
