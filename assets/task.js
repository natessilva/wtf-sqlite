const menuButton = document.getElementById("menu-button");
const contextMenu = document.getElementById("context-menu");
const deleteButton = document.getElementById("delete-button");
const dialog = document.getElementById("delete-dialog");
const taskForm = document.getElementById("task-form");
const descriptionTextarea = document.getElementById("description");
let timeout;

menuButton.addEventListener("click", (e) => {
  e.stopPropagation();
  contextMenu.classList.toggle("active");
});

document.addEventListener("click", () => {
  if (contextMenu.classList.contains("active")) {
    contextMenu.classList.remove("active");
  }
});

contextMenu.addEventListener("click", (e) => {
  e.stopPropagation();
});

deleteButton.addEventListener("click", () => {
  dialog.showModal();
});

dialog.addEventListener("click", (event) => {
  if (event.target === dialog || event.target.classList.contains("cancel")) {
    dialog.close();
  }
});

// auto submit the form
taskForm.addEventListener("input", () => {
  clearTimeout(timeout);
  timeout = setTimeout(() => {
    taskForm.dispatchEvent(new Event("submit", { cancelable: true }));
  }, 200);
});

taskForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const response = await fetch(taskForm.action, {
    method: taskForm.method,
    body: new FormData(taskForm),
  });
  if (!response.ok) {
    const error = await response.text();
    // TODO proper error handling
    console.error("Error:", error);
  }
});

const autoResize = () => {
  descriptionTextarea.style.height = "auto";
  descriptionTextarea.style.height =
    Math.max(120, descriptionTextarea.scrollHeight) + "px";
};
descriptionTextarea.addEventListener("input", autoResize);
autoResize();
