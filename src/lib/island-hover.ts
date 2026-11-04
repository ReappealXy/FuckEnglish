type Scheduler = { set: (fn: () => void, delay: number) => number; clear: (id: number) => void };

// Always cancel the opposite pending action, including before expansion starts.
export function islandHover(schedule: Scheduler, expand: (value: boolean) => void) {
  let timer: number | undefined;
  const cancel = () => { if (timer !== undefined) schedule.clear(timer); timer = undefined; };
  const queue = (value: boolean, delay: number) => {
    cancel();
    timer = schedule.set(() => { timer = undefined; expand(value); }, delay);
  };
  return {
    enter(enabled: boolean, delay: number) { cancel(); if (enabled) queue(true, delay); },
    leave(delay: number) { queue(false, delay); },
    cancel,
  };
}
