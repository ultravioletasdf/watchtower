import type { users } from '../../../../proto/users';

export type Error = {
	incorrectField: string;
	message: string | undefined;
};
export type User = ReturnType<users.User['toObject']>;
