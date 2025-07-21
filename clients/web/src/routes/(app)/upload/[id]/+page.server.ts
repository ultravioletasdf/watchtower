import { videos } from '../../../../server';
import { videos as protoVideos } from '../../../../../../../proto/videos';
import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';
import type { ServerErrorResponse } from '@grpc/grpc-js';

export const actions = {
	default: async ({ cookies, request, params }) => {
		const fd = await request.formData();
		const title = fd.get('title')?.valueOf() as string;
		const description = fd.get('description')?.valueOf() as string;
		const visibility = fd.get('visibility')?.valueOf() as string;
		try {
			const res = await videos.Create(
				new protoVideos.VideosCreateRequest({
					session: cookies.get('session'),
					title,
					description,
					upload_id: parseInt(params.id),
					visibility: visibilityToInt(visibility)
				})
			);
			return res.toObject();
		} catch (e) {
			const err = e as ServerErrorResponse;
			return fail(400, { details: err.details });
		}
	}
} satisfies Actions;

function visibilityToInt(v: string): number {
	switch (v) {
		case 'Public':
			return 0;
		case 'Unlisted':
			return 1;
		case 'Private':
			return 2;
		default:
			return -1;
	}
}
