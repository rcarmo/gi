// Analyse current-run captures before manual disposal. Node/Bun JS sampling
// excludes browser/runtime subprocesses unless those processes captured too.
import { readdirSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';
const root = process.argv[2];
if (!root || !root.startsWith(process.env.GI_TEST_RUN_ROOT + '/')) throw Error('Expected owned current-run captures');
function walk(dir) {
  if (!existsSync(dir)) return;
  for (const entry of readdirSync(dir, {withFileTypes:true})) {
    const path=join(dir,entry.name);
    if (entry.isSymbolicLink()) throw Error('Symlink capture directory');
    if (entry.isDirectory()) {
      if (existsSync(join(path,'cpu.pprof'))) {
        console.log(path);
        for(const [file, extra] of [['cpu.pprof',[]],['mem.pprof',['-sample_index=alloc_space']],['mem.pprof',['-sample_index=alloc_objects']]]) {
          const result=spawnSync(process.env.GO||'go',['tool','pprof','-top','-cum','-nodecount=8',...extra,join(path,file)],{stdio:'inherit'});
          if(result.status!==0)process.exit(result.status||1);
        }
      } else walk(path);
    }
  }
}
walk(root);
