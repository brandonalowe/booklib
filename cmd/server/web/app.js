// Use globals provided by the UMD builds included in index.html
const { createApp, ref, reactive, computed, watchEffect } = Vue
const { createRouter, createWebHistory, useRoute } = VueRouter
const api = {
  async list() { const r = await fetch('/books'); return r.json(); },
  async get(id) { const r = await fetch('/books/' + id); return r.json(); },
  async create(book) { return fetch('/books', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify(book) }); },
  async update(id, book) { return fetch('/books/' + id, { method: 'PUT', headers: {'Content-Type':'application/json'}, body: JSON.stringify(book) }); },
  async delete(id) { return fetch('/books/' + id, { method: 'DELETE' }); }
}

const Home = {
  template: `
    <div class="home">
      <h1>Welcome to BookLib</h1>
      <p>A lightweight personal library manager. Sign in to manage your books.</p>
      <router-link to="/login" class="btn">Sign in</router-link>
    </div>
  `
}

const Login = {
  template: `
    <div class="login">
      <h2>Sign in</h2>
      <p>This is a simulated sign-in for now. Click continue to view your library.</p>
      <router-link to="/library" class="btn">Continue</router-link>
    </div>
  `
}

const Library = {
  setup() {
    const books = ref([])
    const loading = ref(false)
    const selected = ref(null)
    const showAdd = ref(false)

    async function load() {
      loading.value = true
      books.value = await api.list()
      loading.value = false
    }

    function pick(b) {
      selected.value = b
      shelfOpen.value = true
    }

    // wrapper used in template to avoid modifiers causing runtime issues in some builds
    function pickBook(b) { pick(b) }

    async function remove(id) {
      if (!confirm('Delete book?')) return
      await api.delete(id)
      await load()
      selected.value = null
    }

    async function onDeleteSelected() {
      if (!selected.value || !selected.value.id) return
      await remove(selected.value.id)
      closeShelf()
    }

    function openAdd() { showAdd.value = true }
    function closeAdd() { showAdd.value = false }

    const editing = ref(null)
    const editingShelfOpen = ref(false)
    const editForm = reactive({ title: '', author: '', isbn: '', genre: '', read: false })
    const isMobile = ref(false)
    function updateIsMobile() {
      isMobile.value = (typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(max-width:420px)').matches)
    }
    // initialize and attach listener when available
    updateIsMobile()
    if (typeof window !== 'undefined' && window.matchMedia) {
      const mq = window.matchMedia('(max-width:420px)')
      if (mq.addEventListener) mq.addEventListener('change', updateIsMobile)
      else mq.addListener(updateIsMobile)
    }

    function openEdit(b) {
      editing.value = b
      editForm.title = b.title || ''
      editForm.author = b.author || ''
      editForm.isbn = b.isbn || ''
      editForm.genre = b.genre || ''
      editForm.read = !!b.read
      if (isMobile.value) {
        // mobile: close shelf and open the edit-shelf
        shelfOpen.value = false
        editingShelfOpen.value = true
      } else {
        // desktop: close the shelf so the modal overlay is visible and not occluded
        shelfOpen.value = false
        editingShelfOpen.value = false
      }
    }
    function closeEdit() { 
      editing.value = null
      if (editingShelfOpen.value) {
        editingShelfOpen.value = false
        if (selected.value) shelfOpen.value = true
      }
    }
    async function updated() {
      // reload list
      await load()
      // if we were editing a book, update the selected view to the fresh copy
      if (editing && editing.value && editing.value.id) {
        const fresh = books.value.find(b => b.id === editing.value.id)
        if (fresh) selected.value = fresh
      }
      // close any edit UI and return to bookshelf view
      closeEdit()
      editingShelfOpen.value = false
      shelfOpen.value = false
      selected.value = null
    }

    async function created() {
      await load()
      closeAdd()
    }

    async function saveEditMobile() {
      if (!editing.value || !editing.value.id) return
      await api.update(editing.value.id, editForm)
      // reflect the update
      await updated()
    }

  // shelf state for the sliding drawer
  const shelfOpen = ref(false)
  const current = computed(() => (editing.value || selected.value) || { title: '', author: '', genre: '', isbn: '', read: false })
  function closeShelf() { shelfOpen.value = false; selected.value = null; editingShelfOpen.value = false; editing.value = null }

  load()
  return { books, loading, selected, pick, pickBook, remove, onDeleteSelected, showAdd, openAdd, closeAdd, created, editing, openEdit, closeEdit, updated, shelfOpen, closeShelf, editingShelfOpen, editForm, saveEditMobile, isMobile, current }
  },
  template: `
    <div class="library">
      <header class="library-header">
        <h2>Your Library</h2>
        <div>
          <button class="btn" @click="$router.push('/')">Home</button>
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

      <!-- Shelf drawer: slides from right on desktop, bottom on mobile -->
  <transition name="shelf-slide">
    <aside class="shelf" v-if="shelfOpen || editingShelfOpen">
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
            <div v-if="!editingShelfOpen">
              <p><strong>ISBN:</strong> {{ current.isbn }}</p>
              <p><strong>Read:</strong> {{ current.read ? 'Yes' : 'No' }}</p>
              <div class="shelf-actions">
                <button class="btn" @click.stop="openEdit(current)">Edit</button>
                <button class="btn danger" @click.stop="onDeleteSelected">Delete</button>
              </div>
            </div>
            <div v-else class="edit-shelf">
              <label>Title<input v-model="editForm.title"/></label>
              <label>Author<input v-model="editForm.author"/></label>
              <label>ISBN<input v-model="editForm.isbn"/></label>
              <label>Genre<input v-model="editForm.genre"/></label>
              <label><input type="checkbox" v-model="editForm.read"> Read</label>
              <div class="shelf-actions">
                <button class="btn" @click="closeEdit">Cancel</button>
                <button class="btn primary" @click="saveEditMobile">Save</button>
              </div>
            </div>
          </div>
        </aside>
      </transition>

  <add-book-modal v-if="showAdd" @close="closeAdd" @created="created" />
  <edit-book-modal v-if="editing && !isMobile" :book="editing" @close="closeEdit" @updated="updated" />
    </div>
  `
}

const AddBookModal = {
  emits: ['close','created'],
  setup(_, { emit }) {
    const mode = ref('manual') // manual | bulk | scan (scan not implemented)
    const manual = reactive({ title: '', author: '', isbn: '', genre: '', read: false })
    const working = ref(false)

    async function submitManual() {
      working.value = true
      await api.create(manual)
      working.value = false
      emit('created')
    }

    function close() { emit('close') }
    return { mode, manual, submitManual, close, working }
  },
  template: `
    <div class="modal-backdrop">
      <div class="modal">
        <header><h3>Add Book</h3><button class="close" @click="close">×</button></header>
        <div class="modal-body">
          <label><input type="radio" value="manual" v-model="mode"> Manual input</label>
          <label><input type="radio" value="bulk" v-model="mode"> Bulk upload (csv) - coming</label>
          <label><input type="radio" value="scan" v-model="mode"> Scan ISBNs - coming</label>

          <div v-if="mode==='manual'" class="manual-form">
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

          <div v-if="mode==='bulk'">Bulk CSV upload coming soon.</div>
          <div v-if="mode==='scan'">Scan flow coming soon.</div>
        </div>
      </div>
    </div>
  `
}

const Edit = {
  setup() {
    const route = VueRouter.useRoute()
    const id = route.params.id
    const book = reactive({})
    const loading = ref(true)

    async function load() {
      const b = await api.get(id)
      Object.assign(book, b)
      loading.value = false
    }

    async function save() {
      await api.update(id, book)
      alert('Saved')
      window.history.back()
    }

    load()
    return { book, loading, save }
  },
  template: `
    <div>
      <h2>Edit Book</h2>
      <div v-if="loading">Loading...</div>
      <div v-else>
        <input v-model="book.title" />
        <input v-model="book.author" />
        <input v-model="book.isbn" />
        <input v-model="book.genre" />
        <label><input type="checkbox" v-model="book.read"> Read</label>
        <div>
          <button class="btn" @click="save">Save</button>
          <button class="btn" @click="() => window.history.back()">Cancel</button>
        </div>
      </div>
    </div>
  `
}

const routes = [
  { path: '/', component: Home },
  { path: '/login', component: Login },
  { path: '/library', component: Library },
  { path: '/edit/:id', component: Edit },
]

const router = createRouter({ history: createWebHistory(), routes })

// Edit modal component
const EditBookModal = {
  props: ['book'],
  emits: ['close','updated'],
  setup(props, { emit }) {
    const form = reactive({ title: '', author: '', isbn: '', genre: '', read: false })
    const working = ref(false)
    // copy props into form
    function load() {
      form.title = props.book.title || ''
      form.author = props.book.author || ''
      form.isbn = props.book.isbn || ''
      form.genre = props.book.genre || ''
      form.read = !!props.book.read
    }
    async function save() {
      working.value = true
      await api.update(props.book.id, form)
      working.value = false
      emit('updated')
    }
    async function remove() {
      if (!confirm('Delete book?')) return
      working.value = true
      await api.delete(props.book.id)
      working.value = false
      // notify parent to refresh and close shelves
      emit('updated')
    }
    function close() { emit('close') }
    load()
    return { form, working, save, close }
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
            <button class="btn danger" @click="remove" :disabled="working">Delete</button>
            <button class="btn primary" @click="save" :disabled="working">Save</button>
          </div>
        </div>
      </div>
    </div>
  `
}

// Root application
const Root = {
  template: `
    <div>
      <header class="site-header">
        <div class="brand">BookLib</div>
      </header>
      <router-view/>
    </div>`
}

const app = createApp(Root)
app.component('add-book-modal', AddBookModal)
app.component('edit-book-modal', EditBookModal)
app.use(router)
app.mount('#app')
