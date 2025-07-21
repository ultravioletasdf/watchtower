import { thumbnails } from '../../../../../server';
import { videos } from '../../../../../../../../proto/videos';
import { fail, json } from '@sveltejs/kit';

export async function POST({ cookies }) {
	try {
		const res = await thumbnails.CreateUpload(
			new videos.CreateUploadRequest({ session: cookies.get('session') })
		);
		return json(res.toObject());
	} catch (e: unknown) {
		console.log(e);
		throw fail(400);
	}
}
