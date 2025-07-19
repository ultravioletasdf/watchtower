import { serializableGrpcPromise } from '$lib/utils';
import { redirect } from '@sveltejs/kit';
import { users } from '../../../../proto/users';
import { sessions } from '../server';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies, route }) => {
	const session = cookies.get('session');
	console.log(route.id);
	if (session) {
		const user = sessions.GetUser(new users.Session({ token: session }));
		return { streamed: { user: serializableGrpcPromise(user) } };
	} else if (route.id?.startsWith('/(app)')) {
		throw redirect(303, '/sign/in');
	}
};
