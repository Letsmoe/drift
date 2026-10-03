import type { Attachment } from 'svelte/attachments';

/** Attachment that calls onVisible each time the element scrolls into view. */
export function whenVisible(onVisible: () => void): Attachment<HTMLElement> {
	return (element) => {
		const observer = new IntersectionObserver((entries) => {
			if (entries.some((entry) => entry.isIntersecting)) {
				onVisible();
			}
		});
		observer.observe(element);
		return () => observer.disconnect();
	};
}
