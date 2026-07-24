import registry from '../../../shared/agents/registry.json';

export type AgentDef = {
  key: string;
  label: string;
  labelId?: string;
  desc: string;
  inputs: string[];
  outputs: string[];
  envUrl?: string;
  onError?: 'stop' | 'skip' | string;
};

/** Canonical agent registry — single source: shared/agents/registry.json */
export const AGENTS: AgentDef[] = (registry.agents ?? []).map((a) => ({
  key: a.key,
  label: a.label,
  labelId: a.labelId,
  desc: a.desc,
  inputs: a.inputs ?? [],
  outputs: a.outputs ?? [],
  envUrl: a.envUrl,
  onError: a.onError,
}));

export function agentLabel(key: string): string {
  const found = AGENTS.find((a) => a.key === key.toLowerCase());
  return found?.label ?? key.replace(/_/g, ' ');
}
