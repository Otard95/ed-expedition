import './style.css'
import App from './App.svelte'

const app = new App({
  target: document.getElementById('app')
})

// Remove loading screen once app is mounted
const loadingScreen = document.getElementById('loading-screen')
if (loadingScreen) {
  loadingScreen.remove()
}

export default app
