const dialog = document.getElementById("dialog");
const form = dialog.querySelector("form");
const submitButton = form.querySelector('button[type="submit"]');
const addButton = document.getElementById("add-task");
const taskListContainer = document.querySelector(
  ".task-list .scrollable-content"
);

dialog.addEventListener("click", (event) => {
  if (event.target === dialog || event.target.classList.contains("cancel")) {
    dialog.close();
  }
});
addButton.addEventListener("click", () => {
  dialog.showModal();
});

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  submitButton.disabled = true;
  submitButton.textContent = "Submitting...";
  const formData = new FormData(form);
  const response = await fetch("/newTask", {
    method: "POST",
    body: formData,
  });
  if (response.ok) {
    const taskHTML = await response.text();
    taskListContainer.insertAdjacentHTML("beforeend", taskHTML);
    dialog.close();
  } else {
    alert("Failed to create task");
  }
  form.reset();
  submitButton.disabled = false;
  submitButton.textContent = "Submit";

  taskListContainer.scrollTop = taskListContainer.scrollHeight;
});
