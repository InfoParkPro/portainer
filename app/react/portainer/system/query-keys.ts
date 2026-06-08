export const queryKeys = {
  base: () => ['system'] as const,
  selfUpdatePlan: () => [...queryKeys.base(), 'self-update', 'plan'] as const,
};
