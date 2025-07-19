import { users as usersProto } from '../../../proto/users';
import { videos as videosProto } from '../../../proto/videos';
import { credentials } from '@grpc/grpc-js';

export const users = new usersProto.UsersClient('0.0.0.0:50051', credentials.createInsecure());
export const sessions = new usersProto.SessionsClient(
	'0.0.0.0:50051',
	credentials.createInsecure()
);
export const videos = new videosProto.VideosClient('0.0.0.0:50051', credentials.createInsecure());
