/** A route parameter as a string; the router types it as possibly missing. */
export function routeParameter(value: string | undefined): string {
	if (value === undefined) {
		return '';
	}
	return value;
}
