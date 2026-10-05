import net from 'node:net';

const sockets=new Set();
const server=net.createServer(socket=>{
 sockets.add(socket);socket.once('close',()=>sockets.delete(socket));
 socket.write('RFB 003.008\n');socket.on('data',data=>socket.write(data));
});
await new Promise((resolve,reject)=>{server.once('error',reject);server.listen(0,'127.0.0.1',resolve);});
const target=JSON.stringify([{id:'fixture',label:'Disposable RFB fixture',host:'127.0.0.1',port:server.address().port}]);
let child;
const close=()=>{child?.kill();for(const socket of sockets)socket.destroy();server.close();};
process.once('SIGTERM',()=>{close();process.exit(143);});process.once('SIGINT',()=>{close();process.exit(130);});
try{
 child=Bun.spawn(['make','--no-print-directory','GI_TEST_PROFILE_ACTIVE=1','test-ux','PLAYWRIGHT_ARGS=--grep "VNC API"'],{
  env:{...process.env,GI_WEB_VNC_TARGETS:target,GI_WEB_VNC_ALLOW_DIRECT:'false',GI_VNC_PROTOCOL_FIXTURE:'1'},stdout:'inherit',stderr:'inherit',stdin:'ignore',
 });
 const result=await child.exited;
 for(const socket of sockets)socket.destroy();await new Promise(resolve=>server.close(resolve));process.exit(result);
}finally{close();}
