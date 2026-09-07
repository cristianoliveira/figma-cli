import { expect, it } from "vitest";
import { loadWorkspace, saveWorkspace } from "./storage";

it("persists tasks and binary attachments and permits an empty workspace without reseeding", async () => {
  const workspace = {
    projects: ["Personal"],
    tasks: [
      {
        id: "saved",
        title: "Saved task",
        notes: "",
        project: "Personal",
        priority: "low" as const,
        due: "",
        completed: false,
        createdAt: 1,
        attachments: [
          {
            id: "file",
            name: "hello.txt",
            type: "text/plain",
            size: 5,
            blob: new Blob(["hello"]),
          },
        ],
      },
    ],
  };
  await saveWorkspace(workspace);
  const restored = await loadWorkspace();
  expect(restored?.tasks[0].title).toBe("Saved task");
  expect(restored?.tasks[0].attachments[0].name).toBe("hello.txt");
  await saveWorkspace({ tasks: [], projects: [] });
  expect(await loadWorkspace()).toEqual({ tasks: [], projects: [] });
});
