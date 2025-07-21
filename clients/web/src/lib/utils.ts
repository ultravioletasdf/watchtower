import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, 'child'> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };

export function serializableGrpcPromise<T extends { toObject: () => any }>(
	grpcCall: Promise<T>
): Promise<ReturnType<T['toObject']>> {
	return new Promise((res, rej) => {
		grpcCall
			.then((result) => {
				res(result.toObject());
			})
			.catch(rej);
	});
}

export async function upload(e: CustomEvent<any>, endpoint: string): Promise<string> {
	const files: File[] = e.detail.acceptedFiles;
	const file = files[0];
	if (!file) {
		console.log('No file to upload');
	}
	// Allows unlimited uploads atm - FIX THIS
	let res = await fetch(endpoint, { method: 'POST' });
	if (!res.ok) {
		alert('Something went wrong');
		throw "Couldn't get presigned post url";
	}
	const response = (await res.json()) as {
		id: string;
		url: string;
		form_data: Record<string, string>;
	};
	console.log(res, response);
	const formdata = new FormData();
	for (const [k, v] of Object.entries(response.form_data)) {
		formdata.append(k, v);
	}
	formdata.append('file', file);
	res = await fetch(response.url, {
		method: 'POST',
		body: formdata
	});
	return response.id;
}
