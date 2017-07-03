import Vue from 'vue'
import Vuex from 'vuex'
import mutations from './mutations'
import getters from './getters'

Vue.use(Vuex)

const state = {
  user: {},
  req: {},
  baseURL: document.querySelector('meta[name="base"]').getAttribute('content'),
  ssl: (window.location.protocol === 'https:'),
  jwt: '',
  reload: false,
  selected: [],
  multiple: false,
  showInfo: false,
  showHelp: false,
  showDelete: false,
  showRename: false,
  showMove: false,
  showNewFile: false,
  showNewDir: false,
  showDownload: false
}

export default new Vuex.Store({
  strict: process.env.NODE_ENV !== 'production',
  state,
  getters,
  mutations
})
