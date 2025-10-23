export const api = {
  async list() {
    const r = await fetch("/books");
    return r.json();
  },
  async get(id) {
    const r = await fetch("/books/" + id);
    return r.json();
  },
  async create(book) {
    return fetch("/books", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(book),
    });
  },
  async update(id, book) {
    return fetch("/books/" + id, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(book),
    });
  },
  async delete(id) {
    return fetch("/books/" + id, { method: "DELETE" });
  },
};
