import { videos } from '../../../server';
import { videos as protoVideos } from '../../../../../../proto/videos';
import { fail, json } from '@sveltejs/kit';
export async function POST({ cookies }) {
	try {
		const res = await videos.CreateUpload(
			new protoVideos.CreateUploadRequest({ session: cookies.get('session') })
		);
		return json(res.toObject());
	} catch (e) {
		throw fail(400);
	}
}
