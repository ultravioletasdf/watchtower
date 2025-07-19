import { fail, isRedirect, redirect, } from '@sveltejs/kit';
import { sessions } from '../../../server';
import { users } from '../../../../../../proto/users';
import {type ServerErrorResponse, status} from "@grpc/grpc-js"
import type { Actions } from './$types';
import type { Error } from '$lib/types';

const date = new Date()
date.setFullYear(date.getFullYear() + 1)

export const actions = {
	default: async ({request, cookies}) => {
		const form = await  request.formData()
        const email = form.get("email")?.valueOf() as string
        const password = form.get("password")?.valueOf() as string
        console.log(email, password)
        try {
            const session = await sessions.Create(new users.Crededentials({email, password}))
            cookies.set("session", session.token, {
                path: "/",
                httpOnly: true,
                expires: date
            })
            throw redirect(303, "/")
        } catch (e) {
            if (isRedirect(e)) {
                throw e
            }
            let err = e as ServerErrorResponse
            let incorrectField = ""
            console.log(err)
            if (err.message?.includes("email")) {
                incorrectField = "email"
            } else if  (err.message?.includes("password")) {
                incorrectField = "password"
            }
            return fail(400, {error: {incorrectField: incorrectField, message: err.details} as Error})
        }
	}
} satisfies Actions;