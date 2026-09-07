import { useEffect, useState } from "react";
import { Download, FileText, X } from "lucide-react";
import { formatBytes, type Attachment } from "./model";

export default function AttachmentPreview({
  attachment,
  onRemove,
}: {
  attachment: Attachment;
  onRemove?: () => void;
}) {
  const [url, setUrl] = useState("");
  useEffect(() => {
    const objectUrl = URL.createObjectURL(attachment.blob);
    setUrl(objectUrl);
    return () => URL.revokeObjectURL(objectUrl);
  }, [attachment.blob]);
  // Active document formats (HTML, SVG, PDF) are download-only, never embedded.
  const image = [
    "image/png",
    "image/jpeg",
    "image/webp",
    "image/gif",
    "image/avif",
  ].includes(attachment.type);
  const video = [
    "video/mp4",
    "video/webm",
    "video/ogg",
    "video/quicktime",
  ].includes(attachment.type);
  return (
    <div className="attachment">
      <div className={`attachment-preview ${image || video ? "media" : ""}`}>
        {image && url ? (
          <img src={url} alt={attachment.name} />
        ) : video && url ? (
          <video
            src={url}
            controls
            preload="metadata"
            aria-label={`Preview ${attachment.name}`}
          />
        ) : (
          <FileText size={30} strokeWidth={1.3} />
        )}
        {attachment.type === "image/gif" && (
          <span className="gif-badge">GIF</span>
        )}
        {onRemove && (
          <button
            type="button"
            className="remove-file icon-button"
            aria-label={`Remove ${attachment.name}`}
            onClick={onRemove}
          >
            <X size={14} />
          </button>
        )}
      </div>
      <div className="attachment-info">
        <div>
          <strong title={attachment.name}>{attachment.name}</strong>
          <span>{formatBytes(attachment.size)}</span>
        </div>
        <a
          className="icon-button"
          href={url || undefined}
          download={attachment.name}
          aria-label={`Download ${attachment.name}`}
        >
          <Download size={16} />
        </a>
      </div>
    </div>
  );
}
