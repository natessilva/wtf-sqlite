let draggedEl;
let draggedOverEl;
let startX;
let startY;
let startScrollTop;
let dragging = false;
let dragPending = false;
let dragPendingTimer;
let scrollAnimationId;
let scrollSpeed = 10;
let scrollThreshold = 125;
let scrollDirection = 0;
let scrollFactor = 0;
let preventClick = false;
let maxScrollY = 0;
let dx = 0;
let dy = 0;
let dScroll = 0;

const container = document.querySelector(".task-list");
const dragCopy = document.createElement("div");
dragCopy.classList.add("drag-copy");
const parentInput = document.getElementById("parentID");
const addTaskHeader = document.querySelector(".add-task-header");

function scrollAnimation() {
  if (scrollDirection !== 0) {
    const newScroll =
      window.scrollY + scrollDirection * scrollSpeed * scrollFactor;
    if (newScroll > maxScrollY) {
      scrollDirection = 0;
      scrollFactor = 0;
      scrollAnimationId = null;
      return;
    }
    dScroll = startScrollTop + newScroll;
    window.scrollBy(0, scrollDirection * scrollSpeed * scrollFactor);
    dragCopy.style.transform = `translate(${dx}px, ${dy + dScroll}px)`;

    scrollAnimationId = requestAnimationFrame(scrollAnimation);
  }
}

container.addEventListener(
  "dragstart",
  (event) => {
    event.preventDefault();
  },
  { passive: false }
);

container.addEventListener(
  "click",
  (event) => {
    if (preventClick) {
      event.preventDefault();
      preventClick = false;
    }
  },
  { passive: false }
);

container.addEventListener(
  "pointerdown",
  (event) => {
    if (!event.isPrimary) return;
    preventClick = false;
    const target = event.target.closest(".task-item");

    if (target) {
      dragPending = true;
      draggedEl = target;
      startX = event.clientX;
      startY = event.clientY;
      startScrollTop = window.scrollY;
      dragCopy.textContent = target.textContent;

      dragPendingTimer = setTimeout(
        (pointerId) => {
          if (dragPending) {
            container.setPointerCapture(pointerId);
            dragPending = false;
            dragging = true;
            maxScrollY =
              document.documentElement.scrollHeight - window.innerHeight;
            document.body.appendChild(dragCopy);
            dragCopy.style.left = `${startX}px`;
            dragCopy.style.top = `${
              startY + startScrollTop - dragCopy.offsetHeight / 2
            }px`;
          }
        },
        250,
        event.pointerId
      );
    }
  },
  { passive: true }
);

container.addEventListener(
  "pointermove",
  (event) => {
    if (!event.isPrimary) return;
    if (dragPending) {
      const dx = event.clientX - startX;
      const dy = event.clientY - startY;
      if (Math.sqrt(dx * dx + dy * dy) > 5) {
        dragPending = false;
        clearTimeout(dragPendingTimer);
      }
      return;
    }
    if (dragging) {
      dx = event.clientX - startX;
      dy = event.clientY - startY;
      dScroll = window.scrollY - startScrollTop;

      dragCopy.style.transform = `translate(${dx}px, ${dy + dScroll}px)`;

      const distanceFromTop =
        event.clientY - addTaskHeader.getBoundingClientRect().bottom;
      const distanceFromBottom = window.innerHeight - event.clientY;

      if (scrollAnimationId) {
        cancelAnimationFrame(scrollAnimationId);
        scrollAnimationId = null;
      }

      if (distanceFromTop < scrollThreshold && window.scrollY > 0) {
        scrollDirection = -1;
        scrollFactor = 1 - distanceFromTop / scrollThreshold;
        scrollAnimationId = requestAnimationFrame(scrollAnimation);
      } else if (
        distanceFromBottom < scrollThreshold &&
        window.scrollY < maxScrollY
      ) {
        scrollDirection = 1;
        scrollFactor = 1 - distanceFromBottom / scrollThreshold;
        scrollAnimationId = requestAnimationFrame(scrollAnimation);
      } else {
        scrollDirection = 0;
        scrollFactor = 0;
      }

      const target = document
        .elementFromPoint(event.clientX, event.clientY)
        ?.closest(".task-item");
      if (target != null && target != draggedEl) {
        if (draggedOverEl == null || draggedOverEl != target) {
          draggedOverEl?.classList.remove("drag-over");
          draggedOverEl?.classList.remove("drag-over-top");
          draggedOverEl?.classList.remove("drag-over-bottom");
          target.classList.add("drag-over");
          draggedOverEl = target;
          if (dy + dScroll < 0) {
            target.classList.add("drag-over-top");
          } else {
            target.classList.add("drag-over-bottom");
          }
        }
      } else if (
        (target === draggedEl || target == null) &&
        draggedOverEl != null
      ) {
        draggedOverEl.classList.remove("drag-over");
        draggedOverEl.classList.remove("drag-over-top");
        draggedOverEl.classList.remove("drag-over-bottom");
        draggedOverEl = null;
      }
    }
  },
  { passive: true }
);
container.addEventListener(
  "pointerup",
  (event) => {
    if (!event.isPrimary) return;
    if (dragging) {
      preventClick = true;
    }
    container.releasePointerCapture(event.pointerId);
    dragPending = false;
    if (dragPendingTimer) {
      clearTimeout(dragPendingTimer);
    }
    if (dragging) {
      dragging = false;

      if (draggedEl && draggedOverEl) {
        const nextSibling = draggedOverEl.classList.contains("drag-over-top")
          ? draggedOverEl
          : draggedOverEl.nextSibling;
        container.insertBefore(draggedEl, nextSibling);

        const data = new FormData();
        data.append("idToInsert", draggedEl.getAttribute("data-id"));
        if (nextSibling != null) {
          data.append("target", nextSibling.getAttribute("data-id"));
        }
        if (parentInput != null) {
          data.append("parentID", parentInput.value);
        }
        fetch("/insertTaskBefore", {
          method: "POST",
          body: data,
        });
      }
    }

    if (scrollAnimationId) {
      cancelAnimationFrame(scrollAnimationId);
      scrollAnimationId = null;
      scrollDirection = 0;
      scrollFactor = 0;
    }

    if (draggedEl) {
      draggedEl.style.transform = "";
      draggedEl = null;
    }
    if (draggedOverEl) {
      draggedOverEl.classList.remove("drag-over");
      draggedOverEl.classList.remove("drag-over-top");
      draggedOverEl.classList.remove("drag-over-bottom");
      draggedOverEl = null;
    }
    if (dragCopy.parentElement) {
      document.body.removeChild(dragCopy);
      dragCopy.style.transform = "";
    }
  },
  { passive: true }
);

container.addEventListener("touchmove", (event) => {
  if (dragging) {
    event.preventDefault();
  }
});
