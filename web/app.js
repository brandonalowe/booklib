const { createApp, ref, reactive, computed, watchEffect } = Vue;
const { createRouter, createWebHistory, useRoute } = VueRouter;

const TIMEOUT = 50; // ms timeout for focusing inputs in shelves

const api = {
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

// Views
const HomeView = {
  template: `
    <div class="home">
      <h1>Welcome to BookLib</h1>
      <p>A lightweight personal library manager. Sign in to manage your books.</p>
      <router-link to="/login" class="btn">Sign in</router-link>
    </div>
  `,
};

const LoginView = {
  template: `
    <div class="login">
      <h2>Sign in</h2>
      <p>This is a simulated sign-in for now. Click continue to view your library.</p>
      <router-link to="/library" class="btn">Continue</router-link>
    </div>
  `,
};

const LibraryView = {
  setup() {
    const books = ref([]);
    const loading = ref(false);
    const selected = ref(null);
    const isMobile = ref(false);

    const editing = ref(null);
    const editOpen = ref(false);
    const addOpen = ref(false);
    const shelfOpen = ref(false);

    async function load() {
      loading.value = true;
      books.value = await api.list();
      loading.value = false;
    }

    function updateIsMobile() {
      isMobile.value =
        typeof window !== "undefined" &&
        window.matchMedia &&
        window.matchMedia("(max-width:700px)").matches;
    }

    function pick(b) {
      close();
      selected.value = b;
      shelfOpen.value = true;
      console.log("Picked book", b);
    }

    function pickBook(b) {
      pick(b);
    }

    function close() {
      shelfOpen.value = false;
      addOpen.value = false;
      editing.value = editOpen.value = false;
      selected.value = null;
    }

    function openEdit() {
      editing.value = selected.value;

      console.log("openEdit called with book:", editing.value);

      shelfOpen.value = false;
      editOpen.value = true;
      if (isMobile.value) {
        setTimeout(() => {
          const el = document.querySelector(".edit-shelf input");
          if (el) el.focus();
        }, TIMEOUT);
      }
    }

    function closeEdit() {
      editing.value = null;
      if (editOpen.value) {
        editOpen.value = false;
        if (selected.value) shelfOpen.value = true;
      }
    }

    function openAdd() {
      addOpen.value = true;
      if (isMobile.value) {
        setTimeout(() => {
          const el = document.querySelector(".add-shelf input");
          if (el) el.focus();
        }, TIMEOUT);
      }
      console.log("openAdd called");
    }

    async function created() {
      await load();
    }

    async function updated() {
      await load();
      closeEdit();
      editOpen.value = false;
      shelfOpen.value = false;
      selected.value = null;
    }

    async function removed() {
      await load();
      close();
      shelfOpen.value = false;
      selected.value = null;
    }

    updateIsMobile();
    if (typeof window !== "undefined" && window.matchMedia) {
      const mq = window.matchMedia("(max-width:700px)");
      if (mq.addEventListener) mq.addEventListener("change", updateIsMobile);
      else mq.addEventListener(updateIsMobile);
    }
    load();
    return {
      books,
      loading,
      selected,
      isMobile,
      editing,
      editOpen,
      addOpen,
      shelfOpen,
      pick,
      pickBook,
      close,
      openEdit,
      closeEdit,
      openAdd,
      created,
      updated,
      removed,
    };
  },
  template: `
    <div class="library">
        <header class="library-header">
            <h2>Your Library</h2>
            <div>
                <button class="btn primary" @click="openAdd">Add Book</button>
            </div>
        </header>
        <div v-if="loading">Loading...</div>
        <div class="library-grid" v-else>
            <aside class="book-list">
                <ul>
                    <li v-for="b in books" :key="b.id" @click="pickBook(b)" :class="{active: selected && selected.id===b.id}">
                         <div class="title">{{ b.title }}</div>
                         <div class="meta">{{ b.author }} • {{ b.genre }}</div>
                    </li>
                </ul>
            </aside>
        </div>

        <add-book-modal v-if="addOpen && !isMobile" @close="close" @created="created" />
        <add-book-shelf v-if="addOpen && isMobile" @close="close" @created="created" />
        <edit-book-modal v-if="editOpen && !isMobile" :book="editing" @close="closeEdit" @updated="updated" />
        <edit-book-shelf v-if="editOpen && isMobile" :book="editing" @close="closeEdit" @updated="updated" />
        <view-book-shelf v-if="shelfOpen" :book="selected" @close="close" @removed="removed" @editOpen="openEdit" />
    </div>
`,
};

// Components
const AddBooksShelf = {
  emits: ["close", "created"],
  setup(_, { emit }) {
    const addForm = reactive({
      title: "",
      author: "",
      isbn: "",
      genre: "",
      read: false,
    });
    const mode = ref("scan"); // manual | bulk | scan

    const validating = ref(false);
    const searchError = ref("");
    const searchResult = ref(null);
    const added = ref(false);

    async function created() {
      await api.create(addForm);
      emit("created");
    }

    function resetState() {
      addForm.title = "";
      addForm.author = "";
      addForm.isbn = "";
      addForm.genre = "";
      addForm.read = false;
      validating.value = false;
      searchError.value = "";
      searchResult.value = null;
      added.value = false;
      mode.value = "scan";
    }

    function close() {
      resetState();
      emit("close");
    }

    function setMode(m) {
      mode.value = m;
      searchError.value = "";
      searchResult.value = null;
      added.value = false;
    }

    function parseIsbn(v) {
      if (!v) return null;
      const s = String(v).replace(/\D/g, "");
      if (s === "") return null;
      if (s.length > 13) return s.slice(0, 13);
      return s;
    }

    async function doSearch() {
      searchError.value = "";
      searchResult.value = null;
      added.value = false;
      const cleaned = parseIsbn(addForm.isbn);
      if (!cleaned) {
        searchError.value = "ISBN is invalid";
        return;
      }
      validating.value = true;
      try {
        const r = await fetch(`/search/${cleaned}`);
        if (r.status === 400) {
          searchError.value = "ISBN is invalid";
        } else if (r.status === 404) {
          searchError.value = "ISBN returned no results";
        } else if (r.status === 200) {
          const data = await r.json();
          searchResult.value = data;
          // check if book already exists in library (by ISBN digits)
          try {
            const existing = await api.list();
            const norm = String(cleaned).replace(/\D/g, "");
            const found = existing.find((b) => String(b.isbn || "").replace(/\D/g, "") === norm);
            if (found) added.value = true;
          } catch (e) {
            // ignore list errors
          }
        } else {
          searchError.value = `Unexpected response: ${r.status}`;
        }
      } catch (e) {
        searchError.value = "Network error";
      } finally {
        validating.value = false;
      }
    }

    async function addFoundBook() {
      if (!searchResult.value) return;
      const payload = {
        title: searchResult.value.title || "",
        author: searchResult.value.author || "",
        isbn: searchResult.value.isbn || "",
        genre: searchResult.value.genre || "",
        read: false,
      };
      try {
        await api.create(payload);
        added.value = true;
        emit("created");
      } catch (e) {
        searchError.value = "Failed to add book";
      }
    }

    const isScan = computed(() => mode.value === "scan");
    const isManual = computed(() => mode.value === "manual");
    const isBulk = computed(() => mode.value === "bulk");

  // close uses resetState above; ensure cleanup
    return {
      addForm,
      mode,
      setMode,
      isScan,
      isManual,
      isBulk,
      created,
      close,
      doSearch,
      validating,
      searchError,
      searchResult,
      addFoundBook,
      added,
    };
  },
  template: `
  <transition name="shelf-slide-up">
    <aside class="add-shelf">
      <div class="add-shelf-top">
        <button class="shelf-close top" @click="close" aria-label="Close shelf">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
            <path d="M6 9l6 6 6-6" stroke="#d33" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>
      </div>
      <div class="tab-shelf">
        <div class="tab-headers">
          <button :class="['tab-link', { active: isScan }]" @click="setMode('scan')">Search</button>
          <button :class="['tab-link', { active: isManual }]" @click="setMode('manual')">Manual</button>
          <button :class="['tab-link', { active: isBulk }]" @click="setMode('bulk')">Bulk</button>
        </div>
      </div>
      <div class="shelf-body">
          <div v-if="isScan" class="tab-content">
                <div class="search-row">
                  <input class="search-input" v-model="addForm.isbn" placeholder="ISBN" />
                  <button class="search-btn" @click="doSearch" :disabled="validating">Search</button>
                </div>
                <div class="search-feedback">
                  <div v-if="validating">Searching...</div>
                  <div v-if="searchError" class="error">{{ searchError }}</div>
                </div>
                <div class="search-result">
                  <div :class="['search-card', !searchResult ? 'empty' : '']">
                    <div class="cover">
                      <img :src="(searchResult && searchResult.cover_url) || ''" alt="cover" v-if="searchResult && searchResult.cover_url"/>
                      <div v-else>Book Cover</div>
                    </div>
                    <div class="meta">
                      <div class="title">{{ (searchResult && searchResult.title) || 'Title' }}</div>
                      <div class="author">{{ (searchResult && searchResult.author) || 'Author' }}</div>
                      <div class="genre">{{ (searchResult && searchResult.genre) || 'Genre' }}</div>
                      <div class="actions">
                        <button class="btn primary" @click="addFoundBook" v-if="!added">＋ Add</button>
                        <button class="btn success" v-if="added">✓ Added</button>
                      </div>
                    </div>
                  </div>
                </div>
            </div>
      <div v-if="isManual" class="tab-content">
          <input v-model="addForm.title" placeholder="Title"/>
          <input v-model="addForm.author" placeholder="Author"/>
          <input v-model="addForm.isbn" placeholder="ISBN"/>
          <input v-model="addForm.genre" placeholder="Genre"/>
          <input type="checkbox" v-model="addForm.read" /> 
          <div class="shelf-actions">
            <button class="btn" @click="close">Cancel</button>
            <button class="btn primary" @click="created">Add</button>
          </div>
     </div>
     <div v-if="isBulk" class="tab-content">
         <div>Paste CSV rows (title,author,isbn,genre) - coming soon.</div>
     </div>
        </div>
    </aside>
  </transition>
  `,
};

const AddBookModal = {
  emits: ["close", "created"],
  setup(_, { emit }) {
    const mode = ref("scan"); // default to search
    const manual = reactive({
      title: "",
      author: "",
      isbn: "",
      genre: "",
      read: false,
    });
    const working = ref(false);
    const validating = ref(false);
    const searchError = ref("");
    const searchResult = ref(null);
    const added = ref(false);

    async function submitManual() {
      working.value = true;
      await api.create(manual);
      working.value = false;
      emit("created");
    }

    function resetState() {
      manual.title = "";
      manual.author = "";
      manual.isbn = "";
      manual.genre = "";
      manual.read = false;
      validating.value = false;
      searchError.value = "";
      searchResult.value = null;
      added.value = false;
      mode.value = "scan";
    }

    function close() {
      resetState();
      emit("close");
    }

    function setMode(m) {
      mode.value = m;
      searchError.value = "";
      searchResult.value = null;
      added.value = false;
    }

    function parseIsbn(v) {
      if (!v) return null;
      const s = String(v).replace(/\D/g, "");
      if (s === "") return null;
      if (s.length > 13) return s.slice(0, 13);
      return s;
    }

    async function doSearch() {
      searchError.value = "";
      searchResult.value = null;
      added.value = false;
      const cleaned = parseIsbn(manual.isbn);
      if (!cleaned) {
        searchError.value = "ISBN is invalid";
        return;
      }
      validating.value = true;
      try {
        const r = await fetch(`/search/${cleaned}`);
        if (r.status === 400) {
          searchError.value = "ISBN is invalid";
        } else if (r.status === 404) {
          searchError.value = "ISBN returned no results";
        } else if (r.status === 200) {
          const data = await r.json();
          searchResult.value = data;
          // check if book already exists in library (by ISBN digits)
          try {
            const existing = await api.list();
            const norm = String(cleaned).replace(/\D/g, "");
            const found = existing.find((b) => String(b.isbn || "").replace(/\D/g, "") === norm);
            if (found) added.value = true;
          } catch (e) {
            // ignore list errors
          }
        } else {
          searchError.value = `Unexpected response: ${r.status}`;
        }
      } catch (e) {
        searchError.value = "Network error";
      } finally {
        validating.value = false;
      }
    }

    async function addFoundBook() {
      if (!searchResult.value) return;
      const payload = {
        title: searchResult.value.title || "",
        author: searchResult.value.author || "",
        isbn: searchResult.value.isbn || "",
        genre: searchResult.value.genre || "",
        read: false,
      };
      try {
        await api.create(payload);
        added.value = true;
        emit("created");
      } catch (e) {
        searchError.value = "Failed to add book";
      }
    }

    const isScan = computed(() => mode.value === "scan");
    const isManual = computed(() => mode.value === "manual");
    const isBulk = computed(() => mode.value === "bulk");
    return {
      mode,
      manual,
      submitManual,
      close,
      working,
      setMode,
      isScan,
      isManual,
      isBulk,
      validating,
      searchError,
      searchResult,
      doSearch,
      addFoundBook,
      added,
    };
  },
  template: `
    <div class="modal-backdrop">
        <div class="modal">
        <header><h3>Add Book</h3><button class="close" @click="close">×</button></header>
    <div class="modal-body">
      <div class="tab-headers modal-tabs">
        <button :class="['tab-link', { active: mode==='scan' }]" @click="setMode('scan')">Search</button>
        <button :class="['tab-link', { active: mode==='manual' }]" @click="setMode('manual')">Manual</button>
        <button :class="['tab-link', { active: mode==='bulk' }]" @click="setMode('bulk')">Bulk</button>
      </div>
  <div v-if="isScan" class="manual-form">
        <div class="search-row">
          <input class="search-input" v-model="manual.isbn" placeholder="Enter digits only" />
          <button class="search-btn" @click="doSearch" :disabled="validating">Search</button>
        </div>
        <div class="search-feedback">
          <div v-if="validating">Searching...</div>
          <div v-if="searchError" class="error">{{ searchError }}</div>
        </div>
        <div class="search-result modal-result">
          <div :class="['search-card', !searchResult ? 'empty' : '']">
            <div class="cover">
              <img :src="(searchResult && searchResult.cover_url) || ''" alt="cover" v-if="searchResult && searchResult.cover_url"/>
              <div v-else>Book Cover</div>
            </div>
            <div class="meta">
              <div class="title">{{ (searchResult && searchResult.title) || 'Title' }}</div>
              <div class="author">{{ (searchResult && searchResult.author) || 'Author' }}</div>
              <div class="genre">{{ (searchResult && searchResult.genre) || 'Genre' }}</div>
              <div class="modal-actions">
                <button class="btn" @click="doSearch">Search again</button>
                <button class="btn primary" @click="addFoundBook" v-if="!added">＋ Add</button>
                <button class="btn success" v-if="added">✓ Added</button>
              </div>
            </div>
          </div>
        </div>
      </div>
  <div v-if="isManual" class="manual-form">
        <input v-model="manual.title" placeholder="Title" />
        <input v-model="manual.author" placeholder="Author" />
        <input v-model="manual.isbn" placeholder="ISBN" />
        <input v-model="manual.genre" placeholder="Genre" />
        <label><input type="checkbox" v-model="manual.read"> Read</label>
        <div class="modal-actions">
          <button class="btn" @click="close">Cancel</button>
          <button class="btn primary" @click="submitManual" :disabled="working">Add</button>
        </div>
      </div>
  <div v-if="isBulk" class="manual-form">Bulk CSV upload coming soon.</div>
    </div>
        </div>
    </div>
  `,
};

const EditBookShelf = {
  props: ["book"],
  emits: ["close", "updated"],
  setup(props, { emit }) {
    const form = reactive({
      title: "",
      author: "",
      isbn: "",
      genre: "",
      read: false,
    });

    console.log("EditBookShelf props.book:", props.book);

    function load() {
      form.title = props.book.title || "";
      form.author = props.book.author || "";
      form.isbn = props.book.isbn || "";
      form.genre = props.book.genre || "";
      form.read = !!props.book.read;
    }

    async function save() {
      await api.update(props.book.id, form);
      resetForm();
      emit("updated");
    }

    const current = computed(
      () =>
        props.book || {
          title: "",
          author: "",
          genre: "",
          isbn: "",
          read: false,
        },
    );

    function closeEdit() {
      resetForm();
      emit("close");
    }

    function resetForm() {
      form.title = "";
      form.author = "";
      form.isbn = "";
      form.genre = "";
      form.read = false;
    }

    console.log("Loading book form ...");
    load();
    return {
      form,
      current,
      save,
      closeEdit,
    };
  },
  template: `
  <transition name="shelf-slide">
      <aside class="shelf">
          <header class="shelf-header">
              <div>
                  <strong>{{ current.title }}</strong>
                  <div class="meta">{{ current.author }} • {{ current.genre }}</div>
              </div>
              <div>
                  <button class="btn" @click="closeEdit">Close</button>
              </div>
          </header>
          <div class="shelf-body">
              <div class="edit-shelf">
                  <label>Title<input v-model="form.title"/></label>
                  <label>Author<input v-model="form.author"/></label>
                  <label>ISBN<input v-model="form.isbn"/></label>
                  <label>Genre<input v-model="form.genre"/></label>
                  <label><input type="checkbox" v-model="form.read">Read</label>
                  <div class="shelf-actions">
                      <button class="btn" @click="closeEdit">Cancel</button>
                      <button class="btn primary" @click="save">Save</button>
                  </div>
              </div>
          </div>
      </aside>
  </transition>
  `,
};

const EditBookModal = {
  props: ["book"],
  emits: ["close", "updated"],
  setup(props, { emit }) {
    const form = reactive({
      title: "",
      author: "",
      isbn: "",
      genre: "",
      read: false,
    });
    const working = ref(false);
    // copy props into form
    function load() {
      form.title = props.book.title || "";
      form.author = props.book.author || "";
      form.isbn = props.book.isbn || "";
      form.genre = props.book.genre || "";
      form.read = !!props.book.read;
    }
    async function save() {
      working.value = true;
      await api.update(props.book.id, form);
      working.value = false;
      emit("updated");
    }
    function close() {
      emit("close");
    }
    load();
    return { form, working, save, close };
  },
  template: `
    <div class="modal-backdrop">
      <div class="modal">
        <header><h3>Edit Book</h3><button class="close" @click="close">×</button></header>
          <div class="modal-body">
          <input v-model="form.title" placeholder="Title" />
          <input v-model="form.author" placeholder="Author" />
          <input v-model="form.isbn" placeholder="ISBN" />
          <input v-model="form.genre" placeholder="Genre" />
          <label><input type="checkbox" v-model="form.read"> Read</label>
          <div class="modal-actions">
            <button class="btn" @click="close">Cancel</button>
            <button class="btn primary" @click="save" :disabled="working">Save</button>
          </div>
        </div>
      </div>
    </div>
  `,
};

const ViewBookShelf = {
  props: ["book"],
  emits: ["close", "editOpen", "removed"],
  setup(props, { emit }) {
    function closeShelf() {
      emit("close");
    }

    async function remove() {
      if (!confirm("Delete book?")) return;
      await api.delete(props.book.id);
      emit("removed");
    }

    function openEdit() {
      emit("editOpen");
    }

    const current = computed(
      () =>
        props.book || {
          title: "",
          author: "",
          genre: "",
          isbn: "",
          read: false,
        },
    );

    return {
      closeShelf,
      remove,
      openEdit,
      current,
    };
  },
  template: `
  <transition name="shelf-slide">
      <aside class="shelf">
          <header class="shelf-header">
              <div>
                  <strong>{{ current.title }}</strong>
                  <div class="meta">{{ current.author }} • {{ current.genre }}</div>
              </div>
              <div>
                  <button class="btn" @click="closeShelf">Close</button>
              </div>
          </header>
          <div class="shelf-body">
              <p><strong>ISBN:</strong> {{ current.isbn }}</p>
              <p><strong>Read:</strong> {{ current.read ? 'Yes' : 'No' }}</p>
              <div class="shelf-actions">
                  <button class="btn" @click.stop="openEdit">Edit</button>
                  <button class="btn danger" @click.stop="remove">Delete</button>
              </div>
          </div>
      </aside>
  </transition>`,
};

const routes = [
  { path: "/", component: HomeView },
  { path: "/login", component: LoginView },
  { path: "/library", component: LibraryView },
];

const router = createRouter({ history: createWebHistory(), routes });

// Root application
const Root = {
  template: `
    <div>
        <header class="site-header">
            <div>
                <router-link to="/" class="brand">
                    BookLib
                </router-link>
            </div>
        </header>
        <router-view/>
    </div>`,
};

const app = createApp(Root);
app.component("add-book-modal", AddBookModal);
app.component("edit-book-modal", EditBookModal);
app.component("edit-book-shelf", EditBookShelf);
app.component("add-book-shelf", AddBooksShelf);
app.component("view-book-shelf", ViewBookShelf);
app.use(router);
app.mount("#app");
