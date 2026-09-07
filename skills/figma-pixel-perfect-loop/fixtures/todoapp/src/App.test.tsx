import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, expect, it, vi } from "vitest";
import App from "./App";
import { saveWorkspace } from "./storage";
import { seedWorkspace } from "./model";

beforeEach(async () => {
  await saveWorkspace({ tasks: [], projects: ["Personal", "Work"] });
});

it("creates, searches, edits, completes, and deletes a task", async () => {
  const user = userEvent.setup();
  render(<App />);
  await user.click(await screen.findByRole("button", { name: /^New task$/ }));
  await user.type(screen.getByLabelText("Task name"), "Buy coffee");
  await user.click(screen.getByRole("button", { name: "Create task" }));
  await user.click(await screen.findByRole("button", { name: "Buy coffee" }));
  await user.clear(screen.getByLabelText("Task name"));
  await user.type(screen.getByLabelText("Task name"), "Buy good coffee");
  await user.click(screen.getByRole("button", { name: "Save changes" }));
  await screen.findByRole("button", { name: "Buy good coffee" });
  await user.type(screen.getByRole("searchbox"), "missing");
  expect(screen.getByText("No tasks found")).toBeInTheDocument();
  await user.clear(screen.getByRole("searchbox"));
  await user.click(
    screen.getByRole("checkbox", { name: "Complete Buy good coffee" }),
  );
  await user.click(screen.getByRole("button", { name: /^Completed/ }));
  await user.click(
    await screen.findByRole("button", { name: "Buy good coffee" }),
  );
  await user.click(screen.getByRole("button", { name: "Delete task" }));
  await user.click(screen.getByRole("button", { name: "Confirm delete" }));
  await waitFor(() =>
    expect(
      screen.queryByRole("button", { name: "Buy good coffee" }),
    ).not.toBeInTheDocument(),
  );
});

it("uploads a GIF and video, retains them after remount, and removes an attachment", async () => {
  const user = userEvent.setup();
  const app = render(<App />);
  await user.click(await screen.findByRole("button", { name: /^New task$/ }));
  await user.type(screen.getByLabelText("Task name"), "Collect inspiration");
  await user.upload(screen.getByLabelText("Upload attachments"), [
    new File(["gif"], "idea.gif", { type: "image/gif" }),
    new File(["video"], "walkthrough.mp4", { type: "video/mp4" }),
  ]);
  expect(screen.getByAltText("idea.gif")).toBeInTheDocument();
  expect(screen.getByLabelText("Preview walkthrough.mp4")).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Create task" }));
  await screen.findByRole("button", { name: "Collect inspiration" });
  app.unmount();
  render(<App />);
  await user.click(
    await screen.findByRole("button", { name: "Collect inspiration" }),
  );
  const dialog = screen.getByRole("dialog");
  expect(within(dialog).getByAltText("idea.gif")).toBeInTheDocument();
  await user.click(
    within(dialog).getByRole("button", { name: "Remove idea.gif" }),
  );
  await user.click(screen.getByRole("button", { name: "Save changes" }));
  await user.click(
    await screen.findByRole("button", { name: "Collect inspiration" }),
  );
  expect(
    within(screen.getByRole("dialog")).queryByAltText("idea.gif"),
  ).not.toBeInTheDocument();
});

it("creates projects, rejects duplicate names, and filters tasks by project", async () => {
  const user = userEvent.setup();
  render(<App />);
  await user.click(await screen.findByRole("button", { name: "Add project" }));
  await user.type(screen.getByLabelText("Project name"), "work");
  await user.click(screen.getByRole("button", { name: "Create project" }));
  expect(screen.getByRole("alert")).toHaveTextContent("already exists");
  await user.clear(screen.getByLabelText("Project name"));
  await user.type(screen.getByLabelText("Project name"), "Garden");
  await user.click(screen.getByRole("button", { name: "Create project" }));
  expect(
    await screen.findByRole("heading", { name: "Garden" }),
  ).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: /^New task$/ }));
  expect(screen.getByLabelText("Project")).toHaveValue("Garden");
  await user.type(screen.getByLabelText("Task name"), "Plant basil");
  await user.click(screen.getByRole("button", { name: "Create task" }));
  expect(
    await screen.findByRole("button", { name: "Plant basil" }),
  ).toBeInTheDocument();
  await user.click(
    within(screen.getByRole("navigation", { name: "Projects" })).getByRole(
      "button",
      { name: /^Work/ },
    ),
  );
  expect(
    screen.queryByRole("button", { name: "Plant basil" }),
  ).not.toBeInTheDocument();
});

it("sorts by priority, switches layouts, and filters today and upcoming tasks", async () => {
  await saveWorkspace(seedWorkspace());
  const user = userEvent.setup();
  render(<App />);
  await screen.findByRole("button", { name: "Bring the new homepage to life" });
  await user.selectOptions(screen.getByLabelText("Sort tasks"), "priority");
  expect(
    within(screen.getAllByRole("article")[0]).getByRole("button", {
      name: "Bring the new homepage to life",
    }),
  ).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Board view" }));
  expect(screen.getByRole("button", { name: "Board view" })).toHaveAttribute(
    "aria-pressed",
    "true",
  );
  await user.click(screen.getByRole("button", { name: "Today" }));
  expect(screen.getAllByRole("article")).toHaveLength(3);
  await user.click(
    within(
      screen.getByRole("navigation", { name: "Main navigation" }),
    ).getByRole("button", { name: /^Upcoming/ }),
  );
  expect(screen.getAllByRole("article")).toHaveLength(2);
});

it("rejects whitespace-only titles and too many attachments without losing the draft", async () => {
  const user = userEvent.setup();
  render(<App />);
  await user.click(await screen.findByRole("button", { name: /^New task$/ }));
  await user.type(screen.getByLabelText("Task name"), "   ");
  await user.click(screen.getByRole("button", { name: "Create task" }));
  expect(screen.getByRole("alert")).toHaveTextContent("Give your task a name");
  await user.upload(
    screen.getByLabelText("Upload attachments"),
    Array.from({ length: 11 }, (_, i) => new File(["x"], `${i}.txt`)),
  );
  expect(screen.getByRole("alert")).toHaveTextContent("10 files");
  expect(
    screen.queryByRole("link", { name: /Download/ }),
  ).not.toBeInTheDocument();
});

it("lists uploaded documents in Files & media and exposes downloads instead of embedding HTML", async () => {
  const user = userEvent.setup();
  render(<App />);
  await user.click(await screen.findByRole("button", { name: /^New task$/ }));
  await user.type(screen.getByLabelText("Task name"), "Project notes");
  await user.upload(
    screen.getByLabelText("Upload attachments"),
    new File(["<script>alert(1)</script>"], "notes.html", {
      type: "text/html",
    }),
  );
  await user.click(screen.getByRole("button", { name: "Create task" }));
  await screen.findByRole("button", { name: "Project notes" });
  await user.click(screen.getByRole("button", { name: /^Files & media/ }));
  expect(
    screen.getByRole("link", { name: "Download notes.html" }),
  ).toHaveAttribute("download", "notes.html");
  await user.type(screen.getByRole("searchbox"), "missing");
  expect(screen.getByText("No files found")).toBeInTheDocument();
  await user.clear(screen.getByRole("searchbox"));
  await user.click(screen.getByRole("button", { name: "Project notes" }));
  expect(screen.getByLabelText("Task name")).toHaveValue("Project notes");
});

it("reveals a new task when the current search would hide it", async () => {
  const user = userEvent.setup();
  render(<App />);
  await screen.findByRole("button", { name: /^New task$/ });
  await user.type(screen.getByRole("searchbox"), "unrelated search");
  await user.click(screen.getByRole("button", { name: /^New task$/ }));
  await user.type(screen.getByLabelText("Task name"), "Visible new task");
  await user.click(screen.getByRole("button", { name: "Create task" }));
  expect(
    await screen.findByRole("button", { name: "Visible new task" }),
  ).toBeInTheDocument();
  expect(screen.getByRole("searchbox")).toHaveValue("");
});

it("explains local storage and upload limits in help", async () => {
  const user = userEvent.setup();
  render(<App />);
  await user.click(
    await screen.findByRole("button", { name: "A little help" }),
  );
  expect(screen.getByText(/Clearing browser data removes/)).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Close dialog" }));
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
});

it("keeps the draft open and displays a useful error when storage fails", async () => {
  const user = userEvent.setup();
  render(<App />);
  await user.click(await screen.findByRole("button", { name: /^New task$/ }));
  await user.type(screen.getByLabelText("Task name"), "Do not lose me");
  const spy = vi
    .spyOn(IDBDatabase.prototype, "transaction")
    .mockImplementation(() => {
      throw new Error("quota");
    });
  await user.click(screen.getByRole("button", { name: "Create task" }));
  expect(await screen.findByRole("alert")).toHaveTextContent("Could not save");
  expect(screen.getByLabelText("Task name")).toHaveValue("Do not lose me");
  spy.mockRestore();
});
