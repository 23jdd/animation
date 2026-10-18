/// <reference path="../../animation-api.js" />

// 多边形游戏：Go 计算，JavaScript 制作动画，当前引擎导出 GIF。
// 此文件由 Go 嵌入；命令把 DATA 标记替换为真实求解结果，生成独立的 animation.js。
// 生成后的 JS 不需要服务器、浏览器、外部字体或额外数据文件。
// 预览：go run . preview examples/polygon-game/animation.js
// 导出：go run . gif -fps 12 examples/polygon-game/animation.js examples/polygon-game/polygon-game.gif
// 注意：本场景的逻辑画布为 1280 × 720；-scale 只调整导出像素尺寸。

const DATA = /*__POLYGON_DATA__*/null;
const N = DATA.problem.values.length;
const CLEAR = "#00000000";
const P = {
  bg: "#091420", panel: "#102131", border: "#294155", white: "#eaf3fa",
  muted: "#a7bdcf", cyan: "#6dcaff", mint: "#66e3b4", gold: "#ffce78",
  coral: "#ff94aa", cell: "#152b3c", filled: "#1b3548", green: "#173b35",
};

// 时间都来自绝对秒数，不依赖帧数；拖动进度、倒放 seek 和导出均确定一致。
const TIME = { dpBegin: 4.5, cell: 0.22, step: 1.8 };
TIME.dpEnd = TIME.dpBegin + N * (N - 1) * TIME.cell;
TIME.compareEnd = TIME.dpEnd + 2.8;
TIME.mergeBegin = TIME.compareEnd + 2.3;
TIME.mergeEnd = TIME.mergeBegin + (N - 1) * TIME.step;
TIME.end = TIME.mergeEnd + 3.4;

function txt(content, x, y, size = 20, fill = P.white) {
  return text(content, { position: [Math.round(x), Math.round(y)], size, fill, stroke: CLEAR });
}
function box(x, y, width, height, fill = P.panel, stroke = P.border, radius = 10) {
  return rect({position:[x,y],width,height,fill,stroke,strokeWidth:1,radius});
}
function disc(x, y, radius, fill, stroke = CLEAR, strokeWidth = 0) {
  return circle({position:[x,y],radius,fill,stroke,strokeWidth});
}
function rule(a, b, stroke = P.border, strokeWidth = 2) {
  return line({start:a,end:b,stroke,strokeWidth});
}
function at(object, point) { object.x = point[0]; object.y = point[1]; }
function scale(object, factor) { object.scaleX = factor; object.scaleY = factor; }
function clamp(x) { return Math.max(0, Math.min(1, x)); }
function smooth(x) { x = clamp(x); return x*x*(3-2*x); }
function lerp(a, b, t) { return [a[0]+(b[0]-a[0])*t, a[1]+(b[1]-a[1])*t]; }
function op(symbol) { return symbol === "*" ? "×" : "+"; }
function compact(value, limit = 10) {
  if (value.length <= limit) return value;
  const sign = value[0] === "-" ? "−" : "";
  const digits = value[0] === "-" ? value.slice(1) : value;
  return sign + digits[0] + "e" + (digits.length - 1);
}
function operand(value) { return value[0] === "-" ? "("+compact(value)+")" : compact(value); }
function equation(a, symbol, b, value) {
  return operand(a)+" "+op(symbol)+" "+operand(b)+" = "+compact(value);
}
function edgeNames(edges) { return edges.map(edge => "e"+edge).join("、"); }
function same(a, b) { return a.length === b.length && a.every((v, i) => v === b[i]); }
function chainPoint(i, count) {
  const spacing = Math.min(86, 440/Math.max(1,count-1));
  return [-302+(i-(count-1)/2)*spacing, 0];
}

// 所有图形只在加载时创建一次。每帧只更新已有对象的变换和透明度。
const parts = [];
parts.push(box(0,0,1280,720,P.bg,CLEAR,0));
parts.push(txt("多边形游戏",-435,-309,42),txt("POLYGON GAME / INTERVAL DP",-414,-269,17,P.muted));
parts.push(box(417,-301,300,47),txt("Go × 动态规划",417,-301,24,P.mint));
parts.push(txt("负数 · 环形区间 · 最优回溯",417,-259,17,P.muted));
parts.push(box(-302,40,560,484,P.panel,P.border,20),box(302,40,560,484,P.panel,P.border,20));
const chapterNames = ["01  断环成链", "02  区间 DP", "03  最优合并"];
const chapters = chapterNames.map((name, i) => {
  const x = -400+i*400;
  parts.push(box(x,-225,360,39),txt(name,x,-225,20,P.muted));
  const active = group(box(x,-225,360,39,P.green,P.mint),txt(name,x,-225,20,P.mint));
  parts.push(active);
  return active;
});
parts.push(rule([-580,315],[580,315]));
parts.push(txt("Go · 任意精度整数 · 时间 O(n³) / 空间 O(n²)",-286,343,16,P.muted));
parts.push(txt("GO → JAVASCRIPT → GIF",454,343,15,P.mint));
const progress = box(0,315,1160,3,P.mint,CLEAR,0);
parts.push(progress);

// 中文讲解按状态预先建立；不依赖引擎尚未暴露的 text.content / fill 属性。
function captions(title, caption, detail) {
  const g = group(txt(title,-302,-158,23),txt(caption,-302,225,19,P.muted),txt(detail,-302,257,20,P.mint));
  parts.push(g);
  return g;
}
const introCaptions = captions("顶点放整数，边上放运算符","先删一条边，再进行 n−1 次相邻合并","目标：让最后剩下的数最大");
const baseCaptions = captions("断环成链：复制顶点，枚举起点","长度为 1 时：min = max = 顶点值","按照长度 1 → n 逐列求解");
const stateCaptions = DATA.states.map(state => {
  if (state.length === 1) return null;
  const last = (state.start+state.length-2)%N+1;
  return captions("当前区间：v"+state.start+" → v"+last+" · 长度 "+state.length,
    equation(state.left,state.operator,state.right,state.max),
    "min "+compact(state.min)+"     max "+compact(state.max));
});
const compareCaptions = captions("比较所有可能的首删边","最后一列，就是各起点的完整链得分","最优首删边："+edgeNames(DATA.best_cuts));
const unrollCaptions = captions("删掉 e"+DATA.chosen_cut+"，将最优链展开","链起点 v"+DATA.start_vertex+" · 运算顺序由回溯给出","合并不改变剩余顶点的循环顺序");
const stepCaptions = DATA.steps.map((step,i) => captions("第 "+(i+1)+" 步：沿 e"+step.edge+" 合并两个相邻顶点",
  equation(step.left,step.operator,step.right,step.value),(N-i)+" 个顶点 → "+(N-i-1)+" 个顶点"));
const endCaptions = captions("只剩一个顶点 · 全局最优","所有首删边、所有合法合并顺序中的最大值","动态规划求解完成 · 逐步回放校验通过");
const allCaptions = [introCaptions,baseCaptions,...stateCaptions.filter(Boolean),compareCaptions,unrollCaptions,...stepCaptions,endCaptions];

// 圆形顶点、双层高亮和可变换的边；边的两端跟随顶点，运算符始终正立。
function node(value, name, large = false) {
  const radius=large ? 49 : N>6 ? 22 : 29;
  const highlight = disc(0,0,radius+2,CLEAR,P.gold,3);
  const winner = disc(0,0,radius+2,CLEAR,P.mint,3);
  const g = group(disc(0,0,radius,"#173447",P.cyan,2),highlight,winner,
    txt(compact(value,6),0,0,large ? 42 : N>6 ? 22 : 26,value[0] === "-" ? P.coral : P.white),
    txt(name,0,large ? 68 : N>6 ? 36 : 45,N>6 ? 14 : 16,P.muted));
  return {g,highlight,winner,point:[0,0]};
}
function edge(symbol, name) {
  const segment = rule([0,0],[1,0],P.border,3);
  const selected = rule([0,0],[1,0],P.gold,3);
  const badge = group(disc(0,0,N>6 ? 11 : 15,P.panel,P.cyan,1.5),txt(op(symbol),0,0,N>6 ? 17 : 20,P.cyan),txt(name,0,29,14,P.muted));
  return {g:group(segment,selected,badge),segment,selected,badge};
}
function connect(e, a, b) {
  const dx=b[0]-a[0], dy=b[1]-a[1], distance=Math.hypot(dx,dy);
  const margin=Math.min(N>6 ? 23 : 30,distance/2);
  const start=distance>0.001 ? [a[0]+dx/distance*margin,a[1]+dy/distance*margin] : a;
  for (const segment of [e.segment,e.selected]) {
    at(segment,start);segment.rotation=Math.atan2(dy,dx);segment.scaleX=Math.max(0,distance-2*margin);
  }
  at(e.badge,lerp(a,b,0.5));
}
const ringNodes = DATA.problem.values.map((value,i) => node(value,"v"+(i+1)));
const ringEdges = DATA.problem.operators.map((symbol,i) => edge(symbol,"e"+(i+1)));
const ringPoints = ringNodes.map((_,i) => {
  const angle=-Math.PI/2+i*2*Math.PI/N;
  return [Math.round(-302+132*Math.cos(angle)),Math.round(20+132*Math.sin(angle))];
});
const ring = group(...ringEdges.map(e=>e.g),...ringNodes.map(n=>n.g));
parts.push(ring);

const introParts = [txt("游戏规则",302,-155,28)];
["选择任意一条边，免费删除","沿边合并相邻顶点，计算新值","重复合并，直到只剩一个顶点"].forEach((content,i)=>{
  const y=-94+i*57;
  introParts.push(disc(86,y,14,"#1d3a4c"),txt(String(i+1),86,y,16,P.cyan),txt(content,335,y,22));
});
introParts.push(box(302,106,475,58,P.green,CLEAR,12),txt("最后的数越大，得分越高",302,106,25,P.mint));
introParts.push(txt("把环复制为 2n 个顶点，枚举每个起点",302,177,20,P.muted));
introParts.push(txt("关键：负数乘法需要同时保留 min / max",302,219,19,P.gold));
const intro = group(...introParts);
parts.push(intro);

// 表格按 len 从小到大填充。Go 输出的整数是字符串，绝不转成 JS Number 计算。
const tableParts = [txt("起点",76,-78,16,P.muted)];
const normalHeading = group(txt("区间动态规划",302,-155,28),txt("行 = 起点 v_i     列 = 区间长度 len",302,-119,17,P.muted));
const compareHeading = group(txt("全局最优 = 最后一列的最大值",302,-155,24),txt("绿色格子是最优链起点，可能不止一个",302,-119,17,P.muted));
tableParts.push(normalHeading,compareHeading);
const cells = [];
const cw=432/N, rh=246/N;
for (let length=1;length<=N;length++) tableParts.push(txt(String(length),116+(length-0.5)*cw,-78,18,P.muted));
for (let start=1;start<=N;start++) {
  const y=-54+(start-0.5)*rh;
  tableParts.push(txt("v"+start,76,y,18,P.muted));
  for (let length=1;length<=N;length++) {
    const index=(length-1)*N+start-1, x=116+(length-0.5)*cw;
    const base=box(x,y,cw-6,rh-5,P.cell,P.border,7);
    const filled=box(x,y,cw-6,rh-5,P.filled,P.border,7);
    const active=box(x,y,cw-6,rh-5,"#4b3b27",P.gold,7);
    const winner=box(x,y,cw-6,rh-5,"#1b493c",P.mint,7);
    const placeholder=txt("·",x,y,21,P.muted);
    const value=txt(compact(DATA.states[index].max,6),x,y,21);
    cells[index]={filled,active,winner,placeholder,value,start,length};
    tableParts.push(base,filled,active,winner,placeholder,value);
  }
}
const normalFoot = group(txt("每个格子显示该区间的最大值",302,222,18,P.muted),txt("负数乘法：枚举两侧 min / max 的四种组合",302,254,17,P.gold));
const compareFoot = group(txt("首删 "+edgeNames(DATA.best_cuts)+" 均可得到 "+compact(DATA.max_score),302,222,18,P.mint),
  txt("删 e_i 后，从 v_(i+1) 开始（循环编号）",302,254,17,P.muted));
const splitFoot = DATA.states.map(state => state.length===1 ? null :
  txt("切分 k = "+state.split+" · 左侧取 "+state.left_mode+" / 右侧取 "+state.right_mode,302,222,18,P.muted));
tableParts.push(normalFoot,compareFoot,...splitFoot.filter(Boolean));
const table = group(...tableParts);
parts.push(table);

// 最优方案快照完全根据 Go 的回溯步骤建立，不把某组答案写死在动画里。
function snapshot(tokens) {
  const nodes=tokens.map(token=>node(token.value,"v"+token.vertices[0]+(token.vertices.length>1 ? "…v"+token.vertices[token.vertices.length-1] : ""),tokens.length===1));
  const edges=tokens.slice(0,-1).map(token=>{
    const edgeID=token.vertices[token.vertices.length-1];
    return edge(DATA.problem.operators[edgeID-1],"e"+edgeID);
  });
  return {tokens,nodes,edges,g:group(...edges.map(e=>e.g),...nodes.map(n=>n.g))};
}
const initialTokens=Array.from({length:N},(_,i)=>{
  const vertex=(DATA.start_vertex-1+i)%N;
  return {vertices:[vertex+1],value:DATA.problem.values[vertex]};
});
const snapshots=[snapshot(initialTokens)];
const mergeIndices=[];
for (const step of DATA.steps) {
  const previous=snapshots[snapshots.length-1].tokens;
  const index=previous.findIndex((token,i)=>i+1<previous.length && same(token.vertices,step.left_vertices) && same(previous[i+1].vertices,step.right_vertices));
  if (index<0) throw new Error("Go 回溯数据包含非相邻合并");
  const merged={vertices:[...step.left_vertices,...step.right_vertices],value:step.value};
  snapshots.push(snapshot([...previous.slice(0,index),merged,...previous.slice(index+2)]));
  mergeIndices.push(index);
}
parts.push(...snapshots.map(s=>s.g));

const logParts=[txt("回放最优方案",302,-155,28),txt("首删 e"+DATA.chosen_cut+" · 从 v"+DATA.start_vertex+" 开始 · 共 "+(N-1)+" 次合并",302,-115,18,P.muted)];
const logRows=[];
const rowHeight=Math.min(54,266/(N-1));
DATA.steps.forEach((step,i)=>{
  const y=-64+i*rowHeight;
  const completed=box(302,y,484,rowHeight-7,"#183b34","#2a6652",9);
  const active=box(302,y,484,rowHeight-7,"#4b3b27",P.gold,9);
  const content=group(disc(82,y,12,P.border),txt(String(i+1),82,y,15),txt(equation(step.left,step.operator,step.right,step.value),306,y,23),txt("e"+step.edge,523,y,16,P.muted));
  const row=group(box(302,y,484,rowHeight-7,"#172e3f",P.border,9),completed,active,content);
  logRows.push({row,completed,active});logParts.push(row);
});
logParts.push(box(302,238,484,68,P.green,CLEAR,12),txt("最优得分",150,238,23,P.mint),txt(compact(DATA.max_score,16),386,238,36));
const log = group(...logParts);
parts.push(log);
const halo = disc(-302,0,93,CLEAR,P.mint,2);
parts.push(halo);

function resetSnapshot(s) {
  s.nodes.forEach((node,i)=>{node.point=chainPoint(i,s.nodes.length);at(node.g,node.point);scale(node.g,1);node.highlight.opacity=0;node.winner.opacity=0;});
  s.edges.forEach((e,i)=>{connect(e,s.nodes[i].point,s.nodes[i+1].point);e.g.opacity=1;e.selected.opacity=0;});
}
function animateMerge(before,after,index,u) {
  resetSnapshot(before);resetSnapshot(after);
  const travel=smooth((u-0.18)/0.55);
  before.nodes.forEach((node,i)=>{
    const j=i===index+1 ? index : i>index+1 ? i-1 : i;
    node.point=lerp(chainPoint(i,before.nodes.length),chainPoint(j,after.nodes.length),travel);
    at(node.g,node.point);node.highlight.opacity=i===index||i===index+1 ? 1 : 0;
  });
  before.edges.forEach((e,i)=>{
    connect(e,before.nodes[i].point,before.nodes[i+1].point);
    if (i===index) {e.g.opacity=1-smooth((u-0.15)/0.3);e.selected.opacity=1;}
  });
  const appear=smooth((u-0.65)/0.25);
  before.g.opacity=1-appear;after.g.opacity=appear;after.nodes[index].winner.opacity=1;
  if (after.nodes.length===1) scale(after.nodes[0].g,0.7+0.3*appear);
}

// 统一控制器只改变 x/y/rotation/scale/opacity，所有绘制仍由引擎处理。
const controller = group(...parts);
controller.update((self,time)=>{
  const chapter=time>=TIME.compareEnd ? 2 : time>=3 ? 1 : 0;
  chapters.forEach((g,i)=>g.opacity=i===chapter ? 1 : 0);
  const fraction=clamp(time/TIME.end);
  progress.scaleX=fraction;progress.x=-580+580*fraction;
  intro.opacity=1-smooth((time-2.8)/0.7);
  table.opacity=smooth((time-2.8)/0.7)*(1-smooth((time-TIME.mergeBegin+0.6)/0.6));
  log.opacity=smooth((time-TIME.mergeBegin+0.6)/0.6);
  const unroll=smooth((time-TIME.compareEnd)/1.8);
  ring.opacity=time>=TIME.mergeBegin ? 0 : 1;
  const selected=time>=TIME.dpBegin && time<TIME.dpEnd ? N+Math.min(N*(N-1)-1,Math.floor((time-TIME.dpBegin)/TIME.cell)) : -1;
  let completed=time>=3.1 ? N : 0;
  if (time>=TIME.dpBegin) completed=Math.min(N*N,N+1+Math.floor((time-TIME.dpBegin)/TIME.cell));
  if (time>=TIME.dpEnd) completed=N*N;
  cells.forEach((cell,i)=>{
    cell.filled.opacity=i<completed ? 1 : 0;cell.value.opacity=i<completed ? 1 : 0;
    cell.placeholder.opacity=i<completed ? 0 : 1;cell.active.opacity=i===selected ? 1 : 0;
    cell.winner.opacity=time>=TIME.dpEnd && cell.length===N && DATA.best_cuts.some(cut=>cut%N+1===cell.start) ? 1 : 0;
  });
  const comparing=time>=TIME.dpEnd;
  normalHeading.opacity=comparing ? 0 : 1;compareHeading.opacity=comparing ? 1 : 0;
  normalFoot.opacity=!comparing && selected<0 ? 1 : 0;compareFoot.opacity=comparing ? 1 : 0;
  splitFoot.forEach((g,i)=>{if(g)g.opacity=i===selected ? 1 : 0;});
  allCaptions.forEach(g=>g.opacity=0);
  let caption=time<2.8 ? introCaptions : baseCaptions;
  if (selected>=0) caption=stateCaptions[selected];
  if (time>=TIME.dpEnd) caption=compareCaptions;
  if (time>=TIME.compareEnd) caption=unrollCaptions;
  ringNodes.forEach((node,i)=>{
    const chainIndex=(i-(DATA.start_vertex-1)+N)%N;
    node.point=lerp(ringPoints[i],chainPoint(chainIndex,N),unroll);at(node.g,node.point);
    node.g.opacity=smooth((time-0.25-i*0.09)/0.5);node.highlight.opacity=0;node.winner.opacity=0;
    if(selected>=0){const state=DATA.states[selected];const inside=(i-(state.start-1)+N)%N<state.length;node.highlight.opacity=inside?1:0;if(!inside)node.g.opacity=0.3;}
    if(time>=TIME.dpEnd&&i===DATA.start_vertex-1)node.winner.opacity=1;
  });
  ringEdges.forEach((edge,i)=>{
    connect(edge,ringNodes[i].point,ringNodes[(i+1)%N].point);
    edge.g.opacity=smooth((time-0.6-i*0.06)/0.5);edge.selected.opacity=0;
    if(i===DATA.chosen_cut-1&&time>=TIME.compareEnd){edge.selected.opacity=1;edge.g.opacity=1-smooth((time-TIME.compareEnd)/0.65);}
  });
  snapshots.forEach(s=>s.g.opacity=0);
  let activeStep=-1;
  if(time>=TIME.mergeBegin&&time<TIME.mergeEnd){
    activeStep=Math.min(N-2,Math.floor((time-TIME.mergeBegin)/TIME.step));
    const u=(time-TIME.mergeBegin-activeStep*TIME.step)/TIME.step;
    animateMerge(snapshots[activeStep],snapshots[activeStep+1],mergeIndices[activeStep],u);
    caption=stepCaptions[activeStep];
  }
  if(time>=TIME.mergeEnd){const final=snapshots[snapshots.length-1];resetSnapshot(final);final.g.opacity=1;final.nodes[0].winner.opacity=1;caption=endCaptions;}
  caption.opacity=1;
  logRows.forEach((item,i)=>{
    const done=i<activeStep||time>=TIME.mergeEnd;
    item.row.opacity=done||i===activeStep ? 1 : 0.28;item.completed.opacity=done?1:0;item.active.opacity=i===activeStep?1:0;
  });
  halo.opacity=time>=TIME.mergeEnd ? (1-clamp((time-TIME.mergeEnd)/1.8))*0.7 : 0;
  scale(halo,(65+35*smooth((time-TIME.mergeEnd)/1.8))/93);
});
add(controller);
wait(TIME.end);
