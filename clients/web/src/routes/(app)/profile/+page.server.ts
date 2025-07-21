import { videos } from '../../../server.js';
import * as proto from '../../../../../../';
import { serializableGrpcPromise } from '$lib/utils.js';
export function load({ cookies }) {
	const res = videos.GetUserVideos(
		new proto.videos.GetUserVideosRequest({ session: cookies.get('session') })
	);
	return {
		streamed: {
			videos: serializableGrpcPromise(res)
		}
	};
}
