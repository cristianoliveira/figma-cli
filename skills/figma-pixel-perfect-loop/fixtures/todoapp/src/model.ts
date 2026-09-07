export type Priority = "low" | "medium" | "high";
export type View = "all" | "today" | "upcoming" | "completed";
export interface Attachment {
  id: string;
  name: string;
  type: string;
  size: number;
  blob: Blob;
}
export interface Task {
  id: string;
  title: string;
  notes: string;
  project: string;
  priority: Priority;
  due: string;
  completed: boolean;
  attachments: Attachment[];
  createdAt: number;
}
export interface Workspace {
  tasks: Task[];
  projects: string[];
}
export function localDate(date = new Date()): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}
export function filterTasks(
  tasks: Task[],
  options: { view: View; today: string; query?: string; project?: string },
) {
  const query = options.query?.trim().toLowerCase() ?? "";
  return tasks.filter((task) => {
    if (options.project && task.project !== options.project) return false;
    if (
      query &&
      !`${task.title} ${task.notes} ${task.project}`
        .toLowerCase()
        .includes(query)
    )
      return false;
    if (options.view === "completed") return task.completed;
    if (task.completed) return false;
    if (options.view === "today") return task.due === options.today;
    if (options.view === "upcoming") return task.due > options.today;
    return true;
  });
}
export function validateTitle(title: string) {
  if (!title.trim()) return "Give your task a name.";
  if (title.trim().length > 200)
    return "Keep the task name under 201 characters.";
  return "";
}
export function validateFiles(files: File[], existingCount: number) {
  if (files.length + existingCount > 10)
    return "You can attach up to 10 files per task.";
  if (files.some((file) => file.size > 25 * 1024 * 1024))
    return "Each file must be 25 MB or smaller.";
  return "";
}
export function formatBytes(size: number) {
  return size >= 1024 * 1024
    ? `${(size / (1024 * 1024)).toFixed(1)} MB`
    : `${Math.max(1, Math.round(size / 1024))} KB`;
}
export function dateLabel(date: string, today: string) {
  if (!date) return "No due date";
  if (date === today) return "Today";
  const tomorrow = new Date(`${today}T12:00:00`);
  tomorrow.setDate(tomorrow.getDate() + 1);
  if (date === localDate(tomorrow)) return "Tomorrow";
  return new Date(`${date}T12:00:00`).toLocaleDateString("en", {
    month: "short",
    day: "numeric",
  });
}
export function seedWorkspace(now = new Date()): Workspace {
  const today = localDate(now);
  const tomorrow = new Date(now);
  tomorrow.setDate(now.getDate() + 1);
  const later = new Date(now);
  later.setDate(now.getDate() + 3);
  const create = (
    id: string,
    title: string,
    project: string,
    priority: Priority,
    due: string,
    notes: string,
    completed = false,
  ): Task => ({
    id,
    title,
    project,
    priority,
    due,
    notes,
    completed,
    attachments: [],
    createdAt: Number(id),
  });
  return {
    projects: ["Work", "Personal", "Side project"],
    tasks: [
      create(
        "1",
        "Bring the new homepage to life",
        "Work",
        "high",
        today,
        "Explore a fresh direction for the landing page. Think warm colors, thoughtful details, and a little personality.",
      ),
      create(
        "2",
        "A little inspiration collecting",
        "Side project",
        "medium",
        today,
        "Gather the images, GIFs, and little things that spark an idea. Drop your favorites into the attachments.",
      ),
      create(
        "3",
        "Get outside. Take the long way.",
        "Personal",
        "low",
        today,
        "A walk, a good playlist, and absolutely no notifications.",
      ),
      create(
        "4",
        "Record a quick product walkthrough",
        "Work",
        "medium",
        localDate(tomorrow),
        "Keep it simple: the problem, the idea, and the moment it all clicks. Attach the video here when it is ready.",
      ),
      create(
        "5",
        "Plan something just for fun",
        "Personal",
        "low",
        localDate(later),
        "Find a new coffee spot, start a book, or make something with your hands.",
      ),
      create(
        "6",
        "Give that little idea a first draft",
        "Side project",
        "medium",
        "",
        "It does not have to be perfect. It just has to exist.",
      ),
      create(
        "7",
        "Make a little room for the week",
        "Personal",
        "low",
        today,
        "Clear desk, clear head.",
        true,
      ),
      create(
        "8",
        "Send the project proposal",
        "Work",
        "high",
        today,
        "One step closer.",
        true,
      ),
    ],
  };
}
