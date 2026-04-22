// @ts-nocheck
import { html, useEffect } from '../vendor/preact-htm.js';
import { BodyPortal } from './body-portal.js';

/**
 * Image modal for zooming
 */
export function ImageModal({ src, onClose }) {
    useEffect(() => {
        const handleEsc = (e) => {
            if (e.key === 'Escape') onClose();
        };
        document.addEventListener('keydown', handleEsc);
        return () => document.removeEventListener('keydown', handleEsc);
    }, [onClose]);

    return html`
        <${BodyPortal} className="image-modal-portal-root">
            <div class="image-modal" onClick=${onClose}>
                <img src=${src} alt="Full size" />
            </div>
        </${BodyPortal}>
    `;
}
