import {test,expect} from 'bun:test';
import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {parseAuthPolicy} from '../../../web/src/gi-auth-policy.js';
const base={enrolled:true,authenticated:false,totp_enabled:true,mode:'single-user',browser_login_available:true};
test('auth policy fails closed on malformed or unsupported policy',()=>{
 expect(parseAuthPolicy(base)).toEqual(base);
 expect(parseAuthPolicy({...base,enrolled:false,totp_enabled:false})).toMatchObject({enrolled:false});
 for(const value of [null,{},[],{...base,mode:'family'},{...base,authenticated:'true'},{...base,enrolled:'false'},{...base,totp_enabled:false},{...base,browser_login_available:undefined}])expect(()=>parseAuthPolicy(value)).toThrow();
});
test('new passkey contract remains pinned separately from frozen browser inventory',()=>{
 const source=readFileSync('tests/ux/features/additions/piclaw-2026-09-24/piclaw-single-user-passkey-settings.feature');
 expect(createHash('sha256').update(source).digest('hex')).toBe('bd48cab9126778763cab3ddfd8ee04a89bd8c29da62e3bc4188cf24dc3a55b82');
 expect(source.toString().match(/^\s*Scenario(?: Outline)?:/gm)?.length).toBe(26);
});
