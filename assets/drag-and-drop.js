let draggedEl;
let draggedOverEl;
let startX;
let startY;
let startScrollTop;
let dragging = false;
let dragPending = false;
let dragPendingTimer;
let scrollAnimationId;
let scrollSpeed = 15;
let scrollThreshold = 100;
let scrollDirection = 0;
let scrollFactor = 0;
let preventClick = false;

const container = document.querySelector(".task-list .scrollable-content");
const dragCopy = document.createElement("div");
dragCopy.classList.add("drag-copy");

function scrollAnimation() {
  if (scrollDirection !== 0) {
    container.scrollTop += scrollDirection * scrollSpeed * scrollFactor;
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
      startScrollTop = container.scrollTop;
      dragCopy.textContent = target.textContent;

      dragPendingTimer = setTimeout(
        (pointerId) => {
          if (dragPending) {
            container.setPointerCapture(pointerId);
            dragPending = false;
            dragging = true;
            document.body.appendChild(dragCopy);
            dragCopy.style.left = `${startX}px`;
            dragCopy.style.top = `${startY - dragCopy.offsetHeight / 2}px`;
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
      const dx = event.clientX - startX;
      const dy = event.clientY - startY;
      const dScroll = container.scrollTop - startScrollTop;

      dragCopy.style.transform = `translate(${dx}px, ${dy}px)`;

      const containerRect = container.getBoundingClientRect();
      const distanceFromTop = event.clientY - containerRect.top;
      const distanceFromBottom = containerRect.bottom - event.clientY;

      if (scrollAnimationId) {
        cancelAnimationFrame(scrollAnimationId);
        scrollAnimationId = null;
      }

      if (distanceFromTop < scrollThreshold && container.scrollTop > 0) {
        scrollDirection = -1;
        scrollFactor = 1 - distanceFromTop / scrollThreshold;
        scrollAnimationId = requestAnimationFrame(scrollAnimation);
      } else if (
        distanceFromBottom < scrollThreshold &&
        container.scrollTop < container.scrollHeight - container.clientHeight
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
        data.append("target", nextSibling?.getAttribute("data-id"));
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
