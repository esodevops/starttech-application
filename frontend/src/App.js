import React, { useEffect, useState } from 'react';

// The API URL is injected at build time via the REACT_APP_API_URL env variable.
// In development it defaults to http://localhost:8080.
const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

function App() {
  const [todos, setTodos] = useState([]);
  const [title, setTitle] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Fetch all to-dos on first render
  useEffect(() => {
    fetch(`${API_URL}/api/todos`)
      .then((res) => res.json())
      .then((data) => {
        setTodos(data);
        setLoading(false);
      })
      .catch((err) => {
        setError('Could not load todos. Is the backend running?');
        setLoading(false);
      });
  }, []);

  const addTodo = async (e) => {
    e.preventDefault();
    if (!title.trim()) return;

    const res = await fetch(`${API_URL}/api/todos`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title }),
    });
    const newTodo = await res.json();
    setTodos([...todos, newTodo]);
    setTitle('');
  };

  return (
    <div style={{ maxWidth: 600, margin: '40px auto', fontFamily: 'sans-serif' }}>
      <h1>Much To Do ✅</h1>

      <form onSubmit={addTodo} style={{ display: 'flex', gap: 8, marginBottom: 24 }}>
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Add a new todo..."
          style={{ flex: 1, padding: 8, fontSize: 16 }}
        />
        <button type="submit" style={{ padding: '8px 16px', fontSize: 16 }}>
          Add
        </button>
      </form>

      {loading && <p>Loading...</p>}
      {error && <p style={{ color: 'red' }}>{error}</p>}

      <ul style={{ listStyle: 'none', padding: 0 }}>
        {todos.map((todo) => (
          <li
            key={todo.id}
            style={{
              padding: 12,
              borderBottom: '1px solid #eee',
              textDecoration: todo.completed ? 'line-through' : 'none',
            }}
          >
            {todo.title}
          </li>
        ))}
      </ul>
    </div>
  );
}

export default App;
