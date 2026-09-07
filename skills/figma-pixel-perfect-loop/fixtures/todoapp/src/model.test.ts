import { describe, expect, it } from "vitest";
import { filterTasks, validateFiles, validateTitle, type Task } from "./model";

const tasks: Task[] = [
  {
    id: "a",
    title: "Design homepage",
    notes: "Try mint",
    project: "Work",
    priority: "high",
    due: "2026-09-07",
    completed: false,
    attachments: [],
    createdAt: 1,
  },
  {
    id: "b",
    title: "Buy coffee",
    notes: "",
    project: "Personal",
    priority: "low",
    due: "2026-09-08",
    completed: false,
    attachments: [],
    createdAt: 2,
  },
  {
    id: "c",
    title: "Send proposal",
    notes: "",
    project: "Work",
    priority: "medium",
    due: "2026-09-06",
    completed: true,
    attachments: [],
    createdAt: 3,
  },
];
describe("task filtering", () => {
  it("combines case-insensitive search with project and status", () => {
    expect(
      filterTasks(tasks, {
        query: "MINT",
        project: "Work",
        view: "all",
        today: "2026-09-07",
      }).map((t) => t.id),
    ).toEqual(["a"]);
  });
  it("shows only tasks due today in Today and future open tasks in Upcoming", () => {
    expect(
      filterTasks(tasks, { view: "today", today: "2026-09-07" }).map(
        (t) => t.id,
      ),
    ).toEqual(["a"]);
    expect(
      filterTasks(tasks, { view: "upcoming", today: "2026-09-07" }).map(
        (t) => t.id,
      ),
    ).toEqual(["b"]);
  });
  it("shows completed tasks separately and returns empty for unmatched searches", () => {
    expect(
      filterTasks(tasks, { view: "completed", today: "2026-09-07" }).map(
        (t) => t.id,
      ),
    ).toEqual(["c"]);
    expect(
      filterTasks(tasks, {
        view: "all",
        today: "2026-09-07",
        query: "missing",
      }),
    ).toEqual([]);
  });
});
describe("input limits", () => {
  it("rejects blank or oversized titles", () => {
    expect(validateTitle("  ")).toBeTruthy();
    expect(validateTitle("a".repeat(201))).toBeTruthy();
    expect(validateTitle("Plan the week")).toBe("");
  });
  it("accepts files, GIFs, and videos within the limits", () => {
    expect(
      validateFiles(
        [
          new File(["hi"], "notes.txt"),
          new File(["gif"], "fun.gif", { type: "image/gif" }),
          new File(["video"], "demo.mp4", { type: "video/mp4" }),
        ],
        0,
      ),
    ).toBe("");
  });
  it("rejects oversized files and attachment counts over ten without partial acceptance", () => {
    expect(
      validateFiles([{ size: 26 * 1024 * 1024, name: "big.mp4" } as File], 0),
    ).toMatch(/25 MB/);
    expect(validateFiles([new File(["hi"], "notes.txt")], 10)).toMatch(
      /10 files/,
    );
  });
});
