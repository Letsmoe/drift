const recommendContextPrefix = 'recommend:';

/** Human readable name of the context a queue was started from. */
export function contextLabel(context: string): string {
	if (context === '') {
		return '';
	}
	if (context === 'home') {
		return 'Home';
	}
	if (context.startsWith(recommendContextPrefix)) {
		return 'Recommended for you';
	}
	return context;
}
