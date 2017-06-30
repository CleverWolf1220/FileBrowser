const getters = {
  showOverlay: state => {
    return state.showInfo ||
      state.showHelp ||
      state.showDelete ||
      state.showRename ||
      state.showMove ||
      state.showNewFile ||
      state.showNewDir ||
      state.showDownload
  },
  selectedCount: state => state.selected.length
}

export default getters
