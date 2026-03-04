import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'
import { initI18n } from './i18n'

const container = document.getElementById('root')
const root = createRoot(container!)

// i18n baslatildiktan sonra uygulama renderla
initI18n().then(() => {
    root.render(
        <React.StrictMode>
            <App/>
        </React.StrictMode>
    )
})
