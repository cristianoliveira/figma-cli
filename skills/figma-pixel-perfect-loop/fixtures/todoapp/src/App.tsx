import { useEffect, useRef, useState } from "react";
import {
  ArrowDownWideNarrow,
  ArrowRight,
  ArrowUpRight,
  CalendarDays,
  Check,
  CheckCheck,
  ChevronDown,
  ChevronRight,
  CircleHelp,
  Coffee,
  FolderOpen,
  HardDrive,
  LayoutGrid,
  List,
  Menu,
  MoreHorizontal,
  Paperclip,
  Plus,
  Search,
  Sparkles,
  Sprout,
  Sun,
  Sunrise,
  Target,
  X,
} from "lucide-react";
import {
  dateLabel,
  filterTasks,
  localDate,
  seedWorkspace,
  type Task,
  type View,
  type Workspace,
} from "./model";
import { loadWorkspace, saveWorkspace } from "./storage";
import TaskEditor from "./TaskEditor";
import Modal from "./Modal";
import AttachmentPreview from "./AttachmentPreview";

const projectColors = ["mint", "peach", "lavender", "blue", "pink"];
const navigation = [
  { id: "all", label: "All tasks", icon: LayoutGrid },
  { id: "today", label: "My day", icon: Sun },
  { id: "upcoming", label: "Upcoming", icon: CalendarDays },
  { id: "completed", label: "Completed", icon: CheckCheck },
] as const;

function GardenIllustration() {
  return (
    <svg
      className="garden-art"
      viewBox="0 0 380 230"
      fill="none"
      aria-hidden="true"
    >
      <circle cx="248" cy="78" r="43" fill="var(--art-sun)" />
      <path
        d="M248 17V7M306 77h12M291 34l9-9M205 35l-8-9"
        stroke="var(--art-ray)"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <path
        d="M44 206c29-29 53-14 80-43 29-31 47-21 76-38 52-30 101-16 153 60v38H44Z"
        fill="var(--art-hill)"
      />
      <path
        d="M99 220c42-46 70-40 110-56 47-19 82-3 134 38l19 28H99Z"
        fill="var(--art-hill-front)"
      />
      <path
        d="M198 230c23-46 69-44 84-67 12-19-9-29-22-35 32 7 53 23 28 47-21 20-46 29-52 55"
        fill="var(--art-path)"
      />
      <path
        d="M83 211V98"
        stroke="var(--art-stem)"
        strokeWidth="4"
        strokeLinecap="round"
      />
      <path
        d="M83 161c-43 1-55-27-45-48 26-1 45 20 45 48Z"
        fill="var(--art-leaf)"
      />
      <path
        d="M84 137c33-2 43-26 34-44-24 2-34 21-34 44Z"
        fill="var(--art-leaf-light)"
      />
      <path
        d="M83 109c-28-14-29-44-8-57 23 14 26 35 8 57Z"
        fill="var(--art-leaf-dark)"
      />
      <path d="M47 195h72l-8 35H56l-9-35Z" fill="var(--art-pot)" />
      <path d="M43 191h80v11H43z" fill="var(--art-pot-rim)" />
      <rect
        x="148"
        y="39"
        width="68"
        height="83"
        rx="8"
        transform="rotate(-12 148 39)"
        fill="var(--art-paper)"
      />
      <rect
        x="158"
        y="57"
        width="12"
        height="12"
        rx="3"
        transform="rotate(-12 158 57)"
        fill="var(--art-check-bg)"
      />
      <path
        d="m161 61 3 2 4-6M179 55l23-5M176 72l28-6M179 86l19-4"
        stroke="var(--art-check)"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <path
        d="m333 127 3 7 7 3-7 3-3 7-3-7-7-3 7-3 3-7ZM135 18l2 5 5 2-5 2-2 5-2-5-5-2 5-2 2-5Z"
        fill="var(--art-spark)"
      />
      <path
        d="m265 38 4-5 5 4M116 151l5-5 7 3"
        stroke="var(--art-detail)"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}

export default function App() {
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [loadError, setLoadError] = useState(false);
  const [view, setView] = useState<View>("all");
  const [project, setProject] = useState("");
  const [page, setPage] = useState<"tasks" | "files">("tasks");
  const [query, setQuery] = useState("");
  const [layout, setLayout] = useState<"list" | "board">("list");
  const [sort, setSort] = useState("default");
  const [editor, setEditor] = useState<{ task: Task; isNew: boolean } | null>(
    null,
  );
  const [showProject, setShowProject] = useState(false);
  const [projectName, setProjectName] = useState("");
  const [projectError, setProjectError] = useState("");
  const [showHelp, setShowHelp] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const saving = useRef(false);
  const searchRef = useRef<HTMLInputElement>(null);
  const [today, setToday] = useState(localDate());
  useEffect(() => {
    let cancelled = false;
    loadWorkspace()
      .then((data) => {
        if (!cancelled) setWorkspace(data ?? seedWorkspace());
      })
      .catch(() => {
        if (!cancelled) setLoadError(true);
      });
    return () => {
      cancelled = true;
    };
  }, []);
  useEffect(() => {
    const timer = window.setInterval(() => setToday(localDate()), 60_000);
    const shortcut = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === "k") {
        event.preventDefault();
        searchRef.current?.focus();
      }
    };
    window.addEventListener("keydown", shortcut);
    return () => {
      clearInterval(timer);
      window.removeEventListener("keydown", shortcut);
    };
  }, []);
  useEffect(() => {
    if (!notice) return;
    const timer = window.setTimeout(() => setNotice(""), 4500);
    return () => clearTimeout(timer);
  }, [notice]);

  async function persist(next: Workspace) {
    if (saving.current) throw new Error("A save is already in progress.");
    saving.current = true;
    setBusy(true);
    try {
      await saveWorkspace(next);
      setWorkspace(next);
    } finally {
      saving.current = false;
      setBusy(false);
    }
  }
  function go(next: View, nextProject = "") {
    setView(next);
    setProject(nextProject);
    setPage("tasks");
    setSidebarOpen(false);
    setQuery("");
  }
  function newTask() {
    setEditor({
      isNew: true,
      task: {
        id: crypto.randomUUID(),
        title: "",
        notes: "",
        project,
        priority: "medium",
        due: view === "today" ? today : "",
        completed: false,
        attachments: [],
        createdAt: Date.now(),
      },
    });
  }
  async function toggleTask(task: Task) {
    if (!workspace || saving.current) return;
    try {
      await persist({
        ...workspace,
        tasks: workspace.tasks.map((t) =>
          t.id === task.id ? { ...t, completed: !t.completed } : t,
        ),
      });
      setNotice(
        task.completed
          ? "Task moved back to your list."
          : "One less thing. Nicely done!",
      );
    } catch {
      setNotice("Could not save the change. Please try again.");
    }
  }
  async function addProject(event: React.FormEvent) {
    event.preventDefault();
    const name = projectName.trim();
    if (!name) {
      setProjectError("Give your project a name.");
      return;
    }
    if (
      workspace!.projects.some((p) => p.toLowerCase() === name.toLowerCase())
    ) {
      setProjectError("A project with this name already exists.");
      return;
    }
    try {
      await persist({
        ...workspace!,
        projects: [...workspace!.projects, name],
      });
      setShowProject(false);
      setProjectName("");
      go("all", name);
      setNotice("A fresh start. Project created.");
    } catch {
      setProjectError("Could not save the project. Please try again.");
    }
  }
  const tasks = workspace?.tasks ?? [];
  const complete = tasks.filter((t) => t.completed).length;
  const open = tasks.length - complete;
  const todayTasks = tasks.filter(
    (t) => !t.completed && t.due === today,
  ).length;
  const progress = tasks.length
    ? Math.round((complete / tasks.length) * 100)
    : 0;
  const files = tasks
    .flatMap((task) =>
      task.attachments.map((attachment) => ({ task, attachment })),
    )
    .filter(({ task, attachment }) =>
      `${task.title} ${attachment.name}`
        .toLowerCase()
        .includes(query.toLowerCase()),
    );
  const filtered = filterTasks(tasks, { view, today, project, query });
  const sorted = [...filtered].sort((a, b) => {
    if (sort === "priority")
      return (
        { high: 0, medium: 1, low: 2 }[a.priority] -
        { high: 0, medium: 1, low: 2 }[b.priority]
      );
    if (sort === "due") return (a.due || "9999").localeCompare(b.due || "9999");
    if (sort === "newest") return b.createdAt - a.createdAt;
    return a.createdAt - b.createdAt;
  });
  const title =
    page === "files"
      ? "Your little collection"
      : project ||
        (view === "today"
          ? "A day with intention"
          : view === "upcoming"
            ? "Good things ahead"
            : view === "completed"
              ? "Look how far you’ve come"
              : "Make room for what matters.");
  const subtitle =
    page === "files"
      ? "Every file, GIF, and video. Right where you need it."
      : project
        ? "Big ideas happen one small step at a time."
        : view === "completed"
          ? "Every checked box is a little win. Here are yours."
          : "A clear mind starts with a little organization.";
  const projectColor = (name: string) =>
    projectColors[
      Math.max(0, workspace?.projects.indexOf(name) ?? 0) % projectColors.length
    ];

  function taskCard(task: Task) {
    return (
      <article
        className={`task-card ${task.completed ? "is-complete" : ""}`}
        key={task.id}
      >
        <input
          className="task-check"
          type="checkbox"
          checked={task.completed}
          disabled={busy}
          aria-label={`${task.completed ? "Reopen" : "Complete"} ${task.title}`}
          onChange={() => void toggleTask(task)}
        />
        <div className="task-content">
          <div className="task-title-row">
            <button
              className="task-title"
              onClick={() => setEditor({ task, isNew: false })}
            >
              {task.title}
            </button>
            <button
              className="icon-button task-more"
              aria-label={`Edit ${task.title}`}
              onClick={() => setEditor({ task, isNew: false })}
            >
              <MoreHorizontal size={19} />
            </button>
          </div>
          {task.notes && <p className="task-notes">{task.notes}</p>}
          <div className="task-meta">
            {task.project && (
              <button
                className={`project-tag ${projectColor(task.project)}`}
                onClick={() => go("all", task.project)}
              >
                <span className="color-dot" />
                {task.project}
              </button>
            )}
            <span
              className={`due-date ${task.due && task.due < today && !task.completed ? "overdue" : ""}`}
            >
              <CalendarDays size={13} />
              {dateLabel(task.due, today)}
            </span>
            {task.attachments.length > 0 && (
              <span className="file-count">
                <Paperclip size={13} />
                {task.attachments.length}
              </span>
            )}
            <span className={`priority ${task.priority}`}>
              <span />
              {task.priority}
            </span>
          </div>
        </div>
      </article>
    );
  }

  if (loadError)
    return (
      <main className="startup">
        <Sunrise size={40} />
        <h1>Your workspace couldn’t open</h1>
        <p>
          Daylight needs browser storage to keep your tasks and files. Enable
          it, then try again.
        </p>
        <button
          className="button primary"
          onClick={() => window.location.reload()}
        >
          Try again
        </button>
      </main>
    );
  if (!workspace)
    return (
      <main className="startup">
        <Sunrise size={40} />
        <p>Letting a little daylight in…</p>
      </main>
    );
  return (
    <div className="app-shell">
      {sidebarOpen && (
        <button
          className="sidebar-shade"
          aria-label="Close navigation"
          onClick={() => setSidebarOpen(false)}
        />
      )}
      <aside className={`sidebar ${sidebarOpen ? "open" : ""}`}>
        <a
          className="brand"
          href="#"
          onClick={(e) => {
            e.preventDefault();
            go("all");
          }}
        >
          <span className="brand-icon">
            <Sunrise size={25} strokeWidth={1.8} />
          </span>
          daylight<span className="brand-period">.</span>
        </a>
        <div className="workspace-name">
          <span className="workspace-avatar">Y</span>
          <div>
            Your workspace<small>A little space, just for you</small>
          </div>
          <span className="personal-dot" title="Personal workspace" />
        </div>
        <span className="nav-label">WORKSPACE</span>
        <nav aria-label="Main navigation">
          {navigation.map((item) => (
            <button
              key={item.id}
              className={`nav-item ${page === "tasks" && view === item.id && !project ? "active" : ""}`}
              aria-current={
                page === "tasks" && view === item.id && !project
                  ? "page"
                  : undefined
              }
              onClick={() => go(item.id)}
            >
              <item.icon size={19} strokeWidth={1.7} />
              <span>{item.label}</span>
              <span className="nav-count">
                {item.id === "all"
                  ? open
                  : item.id === "today"
                    ? todayTasks
                    : item.id === "completed"
                      ? complete
                      : tasks.filter((t) => !t.completed && t.due > today)
                          .length}
              </span>
            </button>
          ))}
          <button
            className={`nav-item ${page === "files" ? "active" : ""}`}
            onClick={() => {
              setPage("files");
              setProject("");
              setQuery("");
              setSidebarOpen(false);
            }}
          >
            <FolderOpen size={19} strokeWidth={1.7} />
            <span>Files & media</span>
            {tasks.some((t) => t.attachments.length) && (
              <span className="nav-count">
                {tasks.reduce((sum, t) => sum + t.attachments.length, 0)}
              </span>
            )}
          </button>
        </nav>
        <div className="nav-label projects-label">
          YOUR PROJECTS
          <button
            className="icon-button"
            aria-label="Add project"
            onClick={() => {
              setProjectError("");
              setShowProject(true);
            }}
          >
            <Plus size={15} />
          </button>
        </div>
        <nav aria-label="Projects">
          {workspace.projects.map((name) => (
            <button
              key={name}
              className={`nav-item project-nav ${project === name && page === "tasks" ? "active" : ""}`}
              aria-current={
                project === name && page === "tasks" ? "page" : undefined
              }
              onClick={() => go("all", name)}
            >
              <span className={`project-dot ${projectColor(name)}`} />
              <span>{name}</span>
              <span className="nav-count">
                {tasks.filter((t) => t.project === name && !t.completed).length}
              </span>
            </button>
          ))}
          <button
            className="nav-item add-project"
            onClick={() => {
              setProjectError("");
              setShowProject(true);
            }}
          >
            <Plus size={17} />
            New project
          </button>
        </nav>
        <div className="sidebar-bottom">
          <div className="small-reminder">
            <span className="reminder-icon">
              <Sprout size={25} strokeWidth={1.5} />
            </span>
            <strong>Small steps. Real progress.</strong>
            <p>
              You don’t have to do it all.
              <br />
              Just the next little thing.
            </p>
          </div>
          <button
            className="nav-item help-button"
            onClick={() => setShowHelp(true)}
          >
            <CircleHelp size={18} />
            <span>A little help</span>
            <ArrowUpRight size={15} />
          </button>
          <div className="profile">
            <span className="profile-avatar">YO</span>
            <div>
              Your own corner
              <small>
                <span /> Local workspace
              </small>
            </div>
            <button
              className="icon-button"
              aria-label="About your workspace"
              onClick={() => setShowHelp(true)}
            >
              <ChevronDown size={16} />
            </button>
          </div>
        </div>
      </aside>
      <div className="main-shell">
        <header className="topbar">
          <div className="breadcrumb">
            <button
              className="icon-button mobile-menu"
              aria-label="Open navigation"
              onClick={() => setSidebarOpen(true)}
            >
              <Menu size={21} />
            </button>
            <span>Your workspace</span>
            <ChevronRight size={14} />
            <strong>
              {page === "files"
                ? "Files & media"
                : project || navigation.find((n) => n.id === view)?.label}
            </strong>
          </div>
          <div className="topbar-actions">
            <label className="search">
              <Search size={17} />
              <input
                ref={searchRef}
                type="search"
                aria-label="Search tasks and files"
                placeholder={
                  page === "files" ? "Search files…" : "Search anything…"
                }
                value={query}
                onChange={(e) => setQuery(e.target.value)}
              />
              <kbd>⌘ K</kbd>
            </label>
            <span className="header-divider" />
            <span className="mini-avatar" title="Your personal workspace">
              Y
            </span>
          </div>
        </header>
        <main className="main-content">
          <div className="page-heading">
            <div>
              <div className="date-heading">
                <Sun size={14} />
                {new Date(`${today}T12:00:00`).toLocaleDateString("en", {
                  weekday: "long",
                  month: "long",
                  day: "numeric",
                })}
              </div>
              <h1>{title}</h1>
              <p>{subtitle}</p>
            </div>
            <button className="button primary new-task" onClick={newTask}>
              <Plus size={18} />
              New task
            </button>
          </div>
          {page === "tasks" && !project && view === "all" && (
            <section className="welcome-banner">
              <div className="banner-copy">
                <span className="eyebrow">
                  <span /> A FRESH LITTLE START
                </span>
                <h2>
                  A little focus.
                  <br />A lot of possibility.
                </h2>
                <p>
                  Make a plan, find your flow, and leave
                  <br className="desktop-break" /> a little room for the good
                  stuff.
                </p>
                <button onClick={() => go("today")}>
                  Let’s focus on today <ArrowRight size={16} />
                </button>
              </div>
              <GardenIllustration />
              <span className="banner-caption">GROW AT YOUR OWN PACE</span>
            </section>
          )}
          {page === "tasks" && (
            <section className="stats" aria-label="Your task overview">
              <div className="stat">
                <span className="stat-icon mint">
                  <LayoutGrid size={20} strokeWidth={1.7} />
                </span>
                <div>
                  <span>Tasks to do</span>
                  <strong>
                    {open}
                    <small>A little at a time</small>
                  </strong>
                </div>
              </div>
              <div className="stat">
                <span className="stat-icon peach">
                  <Sun size={21} strokeWidth={1.7} />
                </span>
                <div>
                  <span>For today</span>
                  <strong>
                    {todayTasks}
                    <small>You’ve got this</small>
                  </strong>
                </div>
              </div>
              <div className="stat">
                <span className="stat-icon lavender">
                  <CheckCheck size={21} strokeWidth={1.7} />
                </span>
                <div>
                  <span>Completed</span>
                  <strong>
                    {complete}
                    <small>Small wins count</small>
                  </strong>
                </div>
              </div>
              <div className="stat progress-stat">
                <div
                  className="progress-ring"
                  style={{
                    background: `conic-gradient(var(--success) ${progress}%, var(--line) 0)`,
                  }}
                >
                  <span>{progress}%</span>
                </div>
                <div>
                  <span>Your progress</span>
                  <strong className="progress-message">
                    Keep growing <Sprout size={15} />
                  </strong>
                </div>
              </div>
            </section>
          )}
          {page === "tasks" ? (
            <section className="task-section" aria-label="Tasks">
              <div className="section-heading">
                <div>
                  <h2>
                    {project
                      ? "Project tasks"
                      : view === "completed"
                        ? "Your little wins"
                        : "Your tasks"}{" "}
                    <span>{filtered.length}</span>
                  </h2>
                  <p>
                    {view === "completed"
                      ? "Done, dusted, and worth celebrating."
                      : "Everything you need to move a little forward."}
                  </p>
                </div>
                <div className="view-toggle" aria-label="Task layout">
                  <button
                    aria-label="List view"
                    aria-pressed={layout === "list"}
                    onClick={() => setLayout("list")}
                  >
                    <List size={17} />
                  </button>
                  <button
                    aria-label="Board view"
                    aria-pressed={layout === "board"}
                    onClick={() => setLayout("board")}
                  >
                    <LayoutGrid size={16} />
                  </button>
                </div>
              </div>
              <div className="task-toolbar">
                <div className="tabs">
                  {(["all", "today", "upcoming", "completed"] as const).map(
                    (tab) => (
                      <button
                        key={tab}
                        className={view === tab ? "selected" : ""}
                        aria-pressed={view === tab}
                        onClick={() => setView(tab)}
                      >
                        {tab === "all"
                          ? "All tasks"
                          : tab === "today"
                            ? "Today"
                            : tab === "upcoming"
                              ? "Upcoming"
                              : "Done"}
                        {tab === "all" && (
                          <span>
                            {
                              tasks.filter(
                                (t) =>
                                  !t.completed &&
                                  (!project || t.project === project),
                              ).length
                            }
                          </span>
                        )}
                      </button>
                    ),
                  )}
                </div>
                <label className="sort-control">
                  <ArrowDownWideNarrow size={15} />
                  <select
                    aria-label="Sort tasks"
                    value={sort}
                    onChange={(e) => setSort(e.target.value)}
                  >
                    <option value="default">Sort by</option>
                    <option value="due">Due date</option>
                    <option value="priority">Priority</option>
                    <option value="newest">Newest first</option>
                  </select>
                </label>
              </div>
              {sorted.length ? (
                layout === "list" ? (
                  <div className="task-list">{sorted.map(taskCard)}</div>
                ) : (
                  <div className="task-board">
                    {(["high", "medium", "low"] as const).map((priority) => (
                      <div className="board-column" key={priority}>
                        <h3>
                          <span className={`priority ${priority}`}>
                            <span />
                            {priority} priority
                          </span>
                          <small>
                            {
                              sorted.filter((t) => t.priority === priority)
                                .length
                            }
                          </small>
                        </h3>
                        {sorted
                          .filter((t) => t.priority === priority)
                          .map(taskCard)}
                        {!sorted.some((t) => t.priority === priority) && (
                          <p className="empty-column">
                            A little breathing room.
                          </p>
                        )}
                      </div>
                    ))}
                  </div>
                )
              ) : (
                <div className="empty-state">
                  <span>
                    <Coffee size={32} strokeWidth={1.4} />
                  </span>
                  <h3>
                    {query
                      ? "No tasks found"
                      : view === "completed"
                        ? "Your wins will live here"
                        : "A little breathing room."}
                  </h3>
                  <p>
                    {query
                      ? "Try another search or a different filter."
                      : view === "completed"
                        ? "Check off a task and give yourself a little credit."
                        : "Nothing on this list. Add a task, or enjoy the moment."}
                  </p>
                  {query ? (
                    <button
                      className="button secondary"
                      onClick={() => setQuery("")}
                    >
                      Clear search
                    </button>
                  ) : (
                    <button className="button secondary" onClick={newTask}>
                      <Plus size={16} />
                      Add a task
                    </button>
                  )}
                </div>
              )}
              <button className="quick-add" onClick={newTask}>
                <Plus size={17} />
                Add a little something<span>A task, an idea, a next step…</span>
              </button>
            </section>
          ) : (
            <section className="files-section">
              <div className="collection-tip">
                <Paperclip size={22} />
                <div>
                  <strong>More than a to-do list.</strong>
                  <p>
                    Keep your inspiration close. Add files to any task, and find
                    them all here.
                  </p>
                </div>
                <button className="button secondary" onClick={newTask}>
                  Add a task with files
                  <Plus size={16} />
                </button>
              </div>
              {files.length ? (
                <div className="files-grid">
                  {files.map(({ task, attachment }) => (
                    <div key={attachment.id}>
                      <AttachmentPreview attachment={attachment} />
                      <button
                        className="file-task-link"
                        onClick={() => setEditor({ task, isNew: false })}
                      >
                        {task.title}
                        <ArrowUpRight size={13} />
                      </button>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-state">
                  <span>
                    <FolderOpen size={35} strokeWidth={1.3} />
                  </span>
                  <h3>
                    {query ? "No files found" : "A home for your inspiration"}
                  </h3>
                  <p>
                    {query
                      ? "Try a different file or task name."
                      : "Images, GIFs, videos, documents. Attach them to a task to get started."}
                  </p>
                </div>
              )}
            </section>
          )}
          <footer className="page-footer">
            <span>
              <HardDrive size={13} />
              Saved on this device. Just for you.
            </span>
            <span>
              Made for a little more peace of mind <Sun size={13} />
            </span>
          </footer>
        </main>
      </div>
      {notice && (
        <div className="toast" role="status">
          <Check size={17} />
          {notice}
          <button
            className="icon-button"
            aria-label="Dismiss notification"
            onClick={() => setNotice("")}
          >
            <X size={15} />
          </button>
        </div>
      )}
      {editor && (
        <TaskEditor
          key={editor.task.id}
          task={editor.task}
          isNew={editor.isNew}
          projects={workspace.projects}
          onClose={() => setEditor(null)}
          onSave={async (task) => {
            await persist({
              ...workspace,
              tasks: editor.isNew
                ? [...workspace.tasks, task]
                : workspace.tasks.map((t) => (t.id === task.id ? task : t)),
            });
            setEditor(null);
            if (
              page === "tasks" &&
              !filterTasks([task], { view, today, project, query }).length
            ) {
              go(task.completed ? "completed" : "all");
            }
            setNotice(
              editor.isNew
                ? "A little intention, added to your day."
                : "Task updated. All the details, in one place.",
            );
          }}
          onDelete={async () => {
            await persist({
              ...workspace,
              tasks: workspace.tasks.filter((t) => t.id !== editor.task.id),
            });
            setEditor(null);
            setNotice("Task and attachments deleted.");
          }}
        />
      )}
      {showProject && (
        <Modal
          title="A home for your next idea"
          onClose={() => setShowProject(false)}
          busy={busy}
        >
          <form onSubmit={addProject}>
            <label className="field">
              Project name
              <input
                autoFocus
                required
                maxLength={40}
                placeholder="e.g. A fresh start"
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
              />
            </label>
            <p className="muted">
              Keep related tasks together. Make something wonderful.
            </p>
            {projectError && (
              <p role="alert" className="error">
                {projectError}
              </p>
            )}
            <div className="modal-footer">
              <span />
              <button className="button primary" disabled={busy}>
                Create project
                <Plus size={16} />
              </button>
            </div>
          </form>
        </Modal>
      )}
      {showHelp && (
        <Modal
          title="Your own little corner of calm"
          onClose={() => setShowHelp(false)}
        >
          <div className="help-content">
            <div className="help-intro">
              <Sunrise size={32} />
              <p>
                Daylight is a private, local-first space for your tasks and
                everything that goes with them.
              </p>
            </div>
            <h3>
              <Target size={18} />
              Make it yours
            </h3>
            <p>
              Create a task, set a due date, choose a project, and add notes.
              Switch between a list and a priority board. Use ⌘K (or Ctrl+K) to
              search.
            </p>
            <h3>
              <Paperclip size={18} />
              Bring the good stuff
            </h3>
            <p>
              Drop images, animated GIFs, videos, or documents into a task. Up
              to 10 files, 25 MB each. Video playback depends on your browser’s
              supported formats. Other files are available to download.
            </p>
            <h3>
              <HardDrive size={18} />
              Local means local
            </h3>
            <p>
              Tasks and files are stored in this browser using IndexedDB, not
              uploaded to a server. There is no account or cross-device sync.
              Use one tab at a time. Clearing browser data removes your
              workspace, so keep separate copies of important files.
            </p>
            <div className="help-footnote">
              <Sparkles size={17} />
              Start small. You’re already on your way.
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}
