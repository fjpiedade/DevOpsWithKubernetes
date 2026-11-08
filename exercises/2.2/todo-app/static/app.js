const todoForm = document.getElementById("todo-form");
const todoInput = document.getElementById("todo-input");
const todoList = document.getElementById("todo-list");

async function loadTodos() {
  try {
    const response = await fetch("/todos");

    if (!response.ok) {
      throw new Error("Failed to load todos");
    }

    const todos = await response.json();

    todoList.replaceChildren();

    todos.forEach((todo) => {
      const li = document.createElement("li");
      li.textContent = todo.text;
      todoList.appendChild(li);
    });
  } catch (error) {
    console.error("Error loading todos:", error);
    todoList.textContent = "Could not load todos.";
  }
}

todoForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const text = todoInput.value.trim();

  if (!text) {
    return;
  }

  try {
    const response = await fetch("/todos", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ text }),
    });

    if (!response.ok) {
      throw new Error("Failed to create todo");
    }

    todoInput.value = "";

    await loadTodos();
  } catch (error) {
    console.error("Error creating todo:", error);
    alert("Could not create todo.");
  }
});

loadTodos();
