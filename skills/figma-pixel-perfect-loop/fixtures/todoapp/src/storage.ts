import type { Workspace } from "./model";

function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open("daylight-workspace", 1);
    request.onupgradeneeded = () =>
      request.result.createObjectStore("workspace");
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
    request.onblocked = () =>
      reject(new Error("Close other Daylight tabs and try again."));
  });
}

export async function loadWorkspace(): Promise<Workspace | undefined> {
  const db = await openDatabase();
  try {
    return await new Promise<Workspace | undefined>((resolve, reject) => {
      const transaction = db.transaction("workspace", "readonly");
      const request = transaction.objectStore("workspace").get("current");
      transaction.oncomplete = () => resolve(request.result);
      transaction.onerror = () => reject(transaction.error);
      transaction.onabort = () => reject(transaction.error);
    });
  } finally {
    db.close();
  }
}

export async function saveWorkspace(workspace: Workspace): Promise<void> {
  const db = await openDatabase();
  try {
    await new Promise<void>((resolve, reject) => {
      const transaction = db.transaction("workspace", "readwrite");
      transaction.objectStore("workspace").put(workspace, "current");
      transaction.oncomplete = () => resolve();
      transaction.onerror = () => reject(transaction.error);
      transaction.onabort = () => reject(transaction.error);
    });
  } finally {
    db.close();
  }
}
