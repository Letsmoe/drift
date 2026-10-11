// Registers Drift on the local Neoworks stack: the drift organization with the
// developer as its owner, Drift's OAuth client, and every @drift collection
// published through the openschema CLI, which signs in through the browser when
// it has no token. The organization and the client are dev fixtures written
// straight into the database, so the script can run again after a wipe; running
// it twice changes nothing.
//
// The developer is the user with NEOWORKS_EMAIL, or the only user there is.
import { spawnSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { homedir } from 'node:os';
import path from 'node:path';
import { collections, openschemaCli, schemaDirectory, stackPort } from './neoworks-checkout.ts';

const baseDomain = 'neoworks.localhost';
const basePort = stackPort('BASE_PORT', 8443);
const surrealUrl = `http://localhost:${stackPort('SURREAL_PORT', 8000)}`;
const apiUrl = `https://api.${baseDomain}:${basePort}`;
const caddyCertificate = path.join(homedir(), '.local/share/caddy/pki/authorities/local/root.crt');

const clientScopes = [
	'openid',
	'profile',
	'email',
	...collections.flatMap((collection) => [
		`@drift/${collection}:read`,
		`@drift/${collection}:write`
	])
];

// The desktop backend takes the callback on its own loopback port, which the
// OAuth server matches on any port (neoworks-dev/neoworks.dev#43).
const redirectUris = ['http://127.0.0.1/neoworks/callback'];

const organizationPermissions = [
	'organization.manage',
	'members.manage',
	'domains.manage',
	'apps.create',
	'schemas.publish',
	'audit.view',
	'apps.edit',
	'apps.secrets',
	'apps.delete',
	'apps.review',
	'usage.view'
];

const registerStatement = `
	BEGIN TRANSACTION;
	IF $user = NONE { THROW 'no such user' };
	IF (SELECT VALUE handle FROM ONLY $user) = NONE {
		THROW 'the user needs a handle of their own first; claim one in the developer dashboard'
	};
	IF record::exists(handle:drift) {
		IF (SELECT VALUE organization FROM ONLY handle:drift) != organization:drift {
			THROW 'the drift handle belongs to someone else'
		};
	} ELSE {
		CREATE handle:drift SET kind = 'organization', organization = organization:drift;
		CREATE organization:drift SET display_name = 'Drift', created_by = $user;
	};
	IF (SELECT * FROM organization_member WHERE organization = organization:drift AND user = $user) = [] {
		CREATE organization_member SET organization = organization:drift, user = $user, grants = $grants;
	};
	IF !record::exists(client:drift) {
		CREATE client CONTENT { id: 'drift', owner: organization:drift, public: true, first_party: false };
	};
	UPDATE client:drift MERGE {
		name: 'Drift',
		description: 'A YouTube Music player that syncs your library across devices.',
		homepage: 'https://github.com/letsmoe/drift',
		redirect_uris: $redirect_uris,
		scopes: $scopes
	};
	COMMIT TRANSACTION;`;

interface SurrealResult {
	status: string;
	result: unknown;
}

async function surrealQuery(statement: string): Promise<SurrealResult[]> {
	const response = await fetch(`${surrealUrl}/sql`, {
		method: 'POST',
		headers: {
			Accept: 'application/json',
			Authorization: `Basic ${btoa('root:root')}`,
			'surreal-ns': 'neoworks',
			'surreal-db': 'auth'
		},
		body: statement
	});
	if (!response.ok) {
		throw new Error(`SurrealDB answered ${response.status}: ${await response.text()}`);
	}
	return (await response.json()) as SurrealResult[];
}

function throwOnFailure(results: SurrealResult[]): void {
	const failed = results.find((result) => result.status !== 'OK');
	if (failed) {
		throw new Error(`SurrealDB: ${String(failed.result)}`);
	}
}

function surrealString(value: string): string {
	return JSON.stringify(value);
}

async function developerRecord(): Promise<string> {
	const email = process.env.NEOWORKS_EMAIL;
	let statement = 'SELECT VALUE id FROM user;';
	if (email) {
		statement = `SELECT VALUE id FROM user WHERE email = ${surrealString(email)};`;
	}
	const results = await surrealQuery(statement);
	throwOnFailure(results);
	const users = results[0].result as string[];
	if (users.length === 1) {
		return users[0];
	}
	if (email) {
		throw new Error(`no user with the email ${email}`);
	}
	throw new Error(`${users.length} users exist; set NEOWORKS_EMAIL to pick one`);
}

async function register(): Promise<void> {
	const user = await developerRecord();
	const grants = organizationPermissions.map((permission) => ({ permission }));
	const variables = [
		`LET $user = ${user};`,
		`LET $grants = ${JSON.stringify(grants)};`,
		`LET $redirect_uris = ${JSON.stringify(redirectUris)};`,
		`LET $scopes = ${JSON.stringify(clientScopes)};`
	];
	throwOnFailure(await surrealQuery(variables.join('\n') + registerStatement));
	console.log(`drift organization and client registered, owned by ${user}`);
}

function manifestVersion(collection: string): string {
	const manifest = readFileSync(path.join(schemaDirectory, collection, 'openschema.yaml'), 'utf8');
	const match = /^version:\s*(\S+)/m.exec(manifest);
	if (!match) {
		throw new Error(`${collection}/openschema.yaml has no version`);
	}
	return match[1];
}

async function isPublished(collection: string, version: string): Promise<boolean> {
	const url = `${apiUrl}/api/v1/schemas/drift/${collection}/versions/${version}`;
	const response = await fetch(url, { tls: { ca: readFileSync(caddyCertificate, 'utf8') } });
	if (response.status === 404) {
		return false;
	}
	if (!response.ok) {
		throw new Error(`the registry answered ${response.status} for @drift/${collection} ${version}`);
	}
	return true;
}

function publish(collection: string): void {
	const result = spawnSync('bun', [openschemaCli(), 'publish', '--base-port', String(basePort)], {
		cwd: path.join(schemaDirectory, collection),
		stdio: 'inherit',
		env: {
			...process.env,
			OPENSCHEMA_BASE_DOMAIN: baseDomain,
			NODE_EXTRA_CA_CERTS: caddyCertificate
		}
	});
	if (result.status !== 0) {
		process.exit(result.status || 1);
	}
}

async function publishCollections(): Promise<void> {
	for (const collection of collections) {
		const version = manifestVersion(collection);
		if (await isPublished(collection, version)) {
			console.log(`@drift/${collection} ${version} is already published`);
			continue;
		}
		publish(collection);
	}
}

await register();
await publishCollections();
