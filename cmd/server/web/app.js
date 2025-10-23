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
    }

    async function created() {
      await load();
      close();
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
                <router-link to="/" class="btn">Home</router-link>
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
    const mode = ref("manual"); // manual | bulk | scan (scan not implemented)

    async function created() {
      await api.create(addForm);
      emit("created");
    }

  function setMode(m) { mode.value = m }
  function setMode(m) { console.log('AddBookModal setMode ->', m); mode.value = m }
  const isScan = computed(() => mode.value === 'scan');
  const isManual = computed(() => mode.value === 'manual');
  const isBulk = computed(() => mode.value === 'bulk');

    function close() {
      // reset form
      addForm.title = "";
      addForm.author = "";
      addForm.isbn = "";
      addForm.genre = "";
      addForm.read = false;
      emit("close");
    }
    return {
      addForm,
      mode,
      setMode,
  isScan,
  isManual,
  isBulk,
  created,
  close,
    };
  },
  template: `
  <transition name="shelf-slide-up">
    <aside class="add-shelf">
      <div class="tab-shelf">
        <div class="tab-headers">
          <button :class="['tab-link', { active: isScan }]" @click="setMode('scan')">Search</button>
          <button :class="['tab-link', { active: isManual }]" @click="setMode('manual')">Manual</button>
          <button :class="['tab-link', { active: isBulk }]" @click="setMode('bulk')">Bulk</button>
        </div>
      </div>
      <div class="shelf-body">
  <div v-if="isScan" class="tab-content">
          <div class="search-placeholder">Search by ISBN coming soon.</div>
        </div>
  <div v-if="isManual" class="tab-content">
          <label>Title<input v-model="addForm.title" placeholder="Title"/></label>
          <label>Author<input v-model="addForm.author" placeholder="Author"/></label>
          <label>ISBN<input v-model="addForm.isbn" placeholder="ISBN"/></label>
          <label>Genre<input v-model="addForm.genre" placeholder="Genre"/></label>
          <label><input type="checkbox" v-model="addForm.read"> Read</label>
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
    const mode = ref("manual"); // manual | bulk | scan (scan not implemented)
    const manual = reactive({
      title: "",
      author: "",
      isbn: "",
      genre: "",
      read: false,
    });
    const working = ref(false);

    async function submitManual() {
      working.value = true;
      await api.create(manual);
      working.value = false;
      emit("created");
    }

    function close() {
      emit("close");
    }
  // ensure setMode is available
  function setMode(m) { console.log('AddBookModal setMode ->', m); mode.value = m }
  const isScanM = computed(() => mode.value === 'scan');
  const isManualM = computed(() => mode.value === 'manual');
  const isBulkM = computed(() => mode.value === 'bulk');
  return { mode, manual, submitManual, close, working, setMode, isScanM, isManualM, isBulkM };
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
  <div v-if="isScanM" class="manual-form">
        <div class="search-placeholder">Search by ISBN coming soon.</div>
      </div>
  <div v-if="isManualM" class="manual-form">
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
  <div v-if="isBulkM" class="manual-form">Bulk CSV upload coming soon.</div>
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
            <div class="brand">BookLib</div>
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
