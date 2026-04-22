// @ts-nocheck
import { render, useEffect, useLayoutEffect, useState } from '../vendor/preact-htm.js';

export function BodyPortal({ children, className = '' }) {
    const [host, setHost] = useState(null);

    useEffect(() => {
        if (typeof document === 'undefined') return undefined;
        const nextHost = document.createElement('div');
        if (className) nextHost.className = className;
        document.body.appendChild(nextHost);
        setHost(nextHost);
        return () => {
            try {
                render(null, nextHost);
            } finally {
                nextHost.remove();
                setHost((current) => (current === nextHost ? null : current));
            }
        };
    }, [className]);

    useLayoutEffect(() => {
        if (!host) return undefined;
        render(children, host);
        return undefined;
    }, [children, host]);

    return null;
}
