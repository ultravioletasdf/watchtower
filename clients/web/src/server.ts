import { users as usersProto } from '../../../proto/users';
import { videos as videosProto } from '../../../proto/videos';
import { credentials } from '@grpc/grpc-js';

const addr = '0.0.0.0:50051';
const creds = credentials.createInsecure();

export const users = new usersProto.UsersClient(addr, creds);
export const sessions = new usersProto.SessionsClient(addr, creds);
export const videos = new videosProto.VideosClient(addr, creds);
export const thumbnails = new videosProto.ThumbnailsClient(addr, creds);
