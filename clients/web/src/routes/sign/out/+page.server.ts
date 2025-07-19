import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { sessions } from '../../../server';
import { users } from '../../../../../../proto/users';
import { invalidate } from '$app/navigation';

export const load: PageServerLoad = async ({ cookies }) => {
	const session = cookies.get('session');
	if (session) {
		try {
			await sessions.Delete(new users.Session({ token: session }));
		} catch (e) {
			console.log('Failed to delete session: %O', e);
		}
		cookies.delete('session', { path: '/', httpOnly: true, maxAge: -1, expires: new Date() });
	}
	throw redirect(303, '/');
};
