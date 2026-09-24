// Nested horizontal scrollers own their whole gesture, including at an edge.
// Let the browser handle it; never turn a table/code scroll into a session switch.
export function ownsHorizontalGesture(target: Element | null, boundary: Element): boolean {
    for (let element = target; element && element !== boundary; element = element.parentElement) {
        if (element.clientWidth > 0 && element.scrollWidth > element.clientWidth + 1 &&
            /^(auto|scroll)$/.test(getComputedStyle(element).overflowX)) return true;
    }
    return false;
}
