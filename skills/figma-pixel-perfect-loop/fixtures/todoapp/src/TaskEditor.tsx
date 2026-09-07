import { useState, type FormEvent } from "react";
import { ArrowUpRight, CloudUpload, Paperclip, Trash2 } from "lucide-react";
import Modal from "./Modal";
import AttachmentPreview from "./AttachmentPreview";
import {
  validateFiles,
  validateTitle,
  type Task,
  type Priority,
} from "./model";

interface Props {
  task: Task;
  isNew: boolean;
  projects: string[];
  onSave: (task: Task) => Promise<void>;
  onDelete: () => Promise<void>;
  onClose: () => void;
}
export default function TaskEditor({
  task,
  isNew,
  projects,
  onSave,
  onDelete,
  onClose,
}: Props) {
  const [draft, setDraft] = useState(task);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [dragging, setDragging] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const update = (patch: Partial<Task>) =>
    setDraft((current) => ({ ...current, ...patch }));
  function addFiles(files: File[]) {
    const problem = validateFiles(files, draft.attachments.length);
    if (problem) {
      setError(problem);
      return;
    }
    setError("");
    update({
      attachments: [
        ...draft.attachments,
        ...files.map((file) => ({
          id: crypto.randomUUID(),
          name: file.name,
          type: file.type,
          size: file.size,
          blob: file,
        })),
      ],
    });
  }
  async function save(event: FormEvent) {
    event.preventDefault();
    const problem = validateTitle(draft.title);
    if (problem) {
      setError(problem);
      return;
    }
    setBusy(true);
    setError("");
    try {
      await onSave({ ...draft, title: draft.title.trim() });
    } catch {
      setError(
        "Could not save. Browser storage may be full or unavailable. Your draft is still here; try removing a large attachment.",
      );
    } finally {
      setBusy(false);
    }
  }
  async function remove() {
    setBusy(true);
    try {
      await onDelete();
    } catch {
      setError("Could not delete this task. Please try again.");
    } finally {
      setBusy(false);
    }
  }
  return (
    <Modal
      title={isNew ? "Make room for a new task" : "The little details"}
      onClose={onClose}
      busy={busy}
    >
      <form onSubmit={save}>
        <fieldset disabled={busy}>
          <label className="field">
            Task name
            <input
              autoFocus
              value={draft.title}
              onChange={(e) => update({ title: e.target.value })}
              maxLength={200}
              placeholder="What would you like to get done?"
              required
            />
          </label>
          <label className="field">
            Notes <span className="optional">optional</span>
            <textarea
              value={draft.notes}
              onChange={(e) => update({ notes: e.target.value })}
              maxLength={5000}
              rows={3}
              placeholder="An idea, a reminder, or a little more context…"
            />
          </label>
          <div className="form-row">
            <label className="field">
              Project
              <select
                value={draft.project}
                onChange={(e) => update({ project: e.target.value })}
              >
                <option value="">No project</option>
                {projects.map((project) => (
                  <option key={project}>{project}</option>
                ))}
              </select>
            </label>
            <label className="field">
              Due date
              <input
                type="date"
                value={draft.due}
                onChange={(e) => update({ due: e.target.value })}
              />
            </label>
            <label className="field">
              Priority
              <select
                value={draft.priority}
                onChange={(e) =>
                  update({ priority: e.target.value as Priority })
                }
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </label>
          </div>
          <div className="attachment-heading">
            <span>
              <Paperclip size={15} /> Attachments
            </span>
            <small>{draft.attachments.length}/10</small>
          </div>
          <label
            className={`drop-zone ${dragging ? "dragging" : ""}`}
            onDragOver={(e) => {
              e.preventDefault();
              setDragging(true);
            }}
            onDragLeave={() => setDragging(false)}
            onDrop={(e) => {
              e.preventDefault();
              setDragging(false);
              if (!busy) addFiles(Array.from(e.dataTransfer.files));
            }}
          >
            <CloudUpload size={26} strokeWidth={1.5} />
            <strong>Drop a little inspiration here</strong>
            <span>
              or <u>browse files</u> to upload
            </span>
            <small>Images, GIFs, videos & documents · Up to 25 MB each</small>
            <input
              className="visually-hidden"
              type="file"
              multiple
              aria-label="Upload attachments"
              onChange={(e) => {
                addFiles(Array.from(e.target.files ?? []));
                e.target.value = "";
              }}
            />
          </label>
          {draft.attachments.length > 0 && (
            <div className="attachment-grid">
              {draft.attachments.map((file) => (
                <AttachmentPreview
                  key={file.id}
                  attachment={file}
                  onRemove={() =>
                    update({
                      attachments: draft.attachments.filter(
                        (a) => a.id !== file.id,
                      ),
                    })
                  }
                />
              ))}
            </div>
          )}
          {error && (
            <p className="error" role="alert">
              {error}
            </p>
          )}
          <div className="modal-footer">
            {!isNew ? (
              <button
                type="button"
                className="text-button danger"
                onClick={() => setConfirmDelete(true)}
              >
                <Trash2 size={16} />
                Delete task
              </button>
            ) : (
              <span className="local-note">
                Only on your device. Always yours.
              </span>
            )}
            <div>
              <button
                type="button"
                className="button secondary"
                onClick={onClose}
              >
                Cancel
              </button>
              <button className="button primary" type="submit">
                {busy ? "Saving…" : isNew ? "Create task" : "Save changes"}
                <ArrowUpRight size={16} />
              </button>
            </div>
          </div>
          {confirmDelete && (
            <div className="delete-confirm">
              <p>
                Delete this task and all its attachments? This cannot be undone.
              </p>
              <button
                type="button"
                className="button secondary"
                onClick={() => setConfirmDelete(false)}
              >
                Keep task
              </button>
              <button
                type="button"
                className="button destructive"
                onClick={remove}
              >
                Confirm delete
              </button>
            </div>
          )}
        </fieldset>
      </form>
    </Modal>
  );
}
