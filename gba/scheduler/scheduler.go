// https://www.gregorygaines.com/blog/emulator-polling-vs-scheduler-game-loop/ - Emulator polling vs Scheduler Game Loop
// Scheduler is responsible for timing and sync of the emulator
// For each event added to the scheduler, execute it based on the current time vs the scheduled time
// If event is late, execute it immediately
// TODO: Decide how to handle late events to ensure proper sync
package scheduler

