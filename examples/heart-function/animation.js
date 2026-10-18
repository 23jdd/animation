/// <reference path="../../animation-api.js" />

// 《把一个心形，写成函数》 / THE QUIET SCIENCE 02 / 44 seconds.
// JS 描述全部画面与动作；当前 Go 引擎负责预览、PNG、GIF 和 MP4 渲染。
// 预览：go run . preview examples/heart-function/animation.js
// 视频：go run . mp4 -fps 30 examples/heart-function/animation.js examples/heart-function/heart-function.mp4
// GIF：go run . gif -fps 12 -scale 0.75 examples/heart-function/animation.js examples/heart-function/preview.gif
// 1280×720，中文字幕，无音乐、无配音，不使用图片贴图。

const FILM = {duration:44, drawStart:12, drawEnd:28};
const TAU = Math.PI*2;
const SAMPLES = 1600;
const CLEAR = "#00000000";
const P = {
  paper:"#f7f2ec", ink:"#423942", muted:"#847984", faint:"#a99ba2",
  rose:"#bd5575", blush:"#edd3dc", line:"#ddd2d4", gold:"#b39a72", grid:"#ebe2df",
};

function label(content,x,y,size=22,fill=P.ink) {
  return text(content,{position:[Math.round(x),Math.round(y)],size,fill,stroke:CLEAR});
}
function dot(x,y,radius,fill,stroke=CLEAR,strokeWidth=0) {
  return circle({position:[x,y],radius,fill,stroke,strokeWidth});
}
function segment(a,b,stroke=P.line,strokeWidth=1) {
  return line({start:a,end:b,stroke,strokeWidth});
}
function place(object,point) {object.x=point[0];object.y=point[1];}
function clamp(x) {return Math.max(0,Math.min(1,x));}
function smooth(x) {x=clamp(x);return x*x*(3-2*x);}
function fade(time,start,end,seconds=.65) {
  return (start===0?1:smooth((time-start)/seconds))*(1-smooth((time-end+seconds)/seconds));
}

/** The mathematical curve: positive y is UP, and t is measured in radians.
 * @param {number} t @returns {Vec2}
 */
function heart(t) {
  return [
    16*Math.pow(Math.sin(t),3),
    13*Math.cos(t)-5*Math.cos(2*t)-2*Math.cos(3*t)-Math.cos(4*t),
  ];
}

const POINTS = Array.from({length:SAMPLES+1},(_,i)=>heart(TAU*i/SAMPLES));
const BOUNDS = POINTS.reduce((b,p)=>({
  minX:Math.min(b.minX,p[0]),maxX:Math.max(b.maxX,p[0]),
  minY:Math.min(b.minY,p[1]),maxY:Math.max(b.maxY,p[1]),
}),{minX:Infinity,maxX:-Infinity,minY:Infinity,maxY:-Infinity});

// Preserve the aspect ratio; x and y use the same number of pixels per unit.
function fitScale(width,height) {
  return Math.min(width/(BOUNDS.maxX-BOUNDS.minX),height/(BOUNDS.maxY-BOUNDS.minY));
}
function screenPoint(t,scale) {
  const p=heart(t);return [p[0]*scale,-p[1]*scale];
}
function curvePath(points,stroke=P.rose,width=2.5,fill=CLEAR,closed=false) {
  const commands=points.map((p,i)=>[i===0?"M":"L",p[0],p[1]]);
  if(closed)commands.push(["Z"]);
  return path({commands,stroke,strokeWidth:width,fill});
}
function heartArt(scale,fillOpacity=.45) {
  const points=POINTS.map(p=>[p[0]*scale,-p[1]*scale]);
  const fill=curvePath(points,CLEAR,0,P.blush,true);fill.opacity=fillOpacity;
  const soft=curvePath(points,P.rose,9);soft.opacity=.065;
  const outline=curvePath(points,P.rose,2.5);
  return {g:group(fill,soft,outline),fill,outline};
}
function centerArt(object,center,scale) {
  object.x=center[0]-(BOUNDS.minX+BOUNDS.maxX)*scale/2;
  object.y=center[1]+(BOUNDS.minY+BOUNDS.maxY)*scale/2;
}

// A canonical unit line is transformed to connect any two points.
// This only changes an existing object; no drawing objects are created per frame.
function connect(object,a,b) {
  const dx=b[0]-a[0],dy=b[1]-a[1],length=Math.hypot(dx,dy);
  place(object,a);object.rotation=Math.atan2(dy,dx);object.scaleX=length;
  object.opacity=length>1e-9?1:0;
}

// Reveal by PARAMETER t, not by arc length, so the drawn head and heart(t) agree.
// Completed 32-sample blocks use one path each; just 32 reusable lines draw the
// current partial block. This avoids painting 1600 individual lines every frame.
function makeTracer(scale) {
  const blockSize=32;
  const points=POINTS.map(p=>[p[0]*scale,-p[1]*scale]);
  const blocks=[];
  for(let start=0;start<SAMPLES;start+=blockSize) {
    blocks.push(curvePath(points.slice(start,start+blockSize+1),P.rose,2.8));
  }
  const partial=Array.from({length:blockSize},()=>segment([0,0],[1,0],P.rose,2.8));
  const tracer={g:group(...blocks,...partial),blocks,partial,t:0,head:[0,0]};
  tracer.setParameter=t=>{
    t=Math.max(0,Math.min(TAU,t));
    const index=Math.min(SAMPLES,Math.floor(t/TAU*SAMPLES));
    const block=Math.floor(index/blockSize),first=block*blockSize;
    blocks.forEach((item,i)=>{item.opacity=i<block?1:0;});
    const head=screenPoint(t,scale);
    partial.forEach((item,j)=>{
      const i=first+j;
      if(i>=SAMPLES||i>index){item.opacity=0;return;}
      connect(item,points[i],i===index?head:points[i+1]);
    });
    tracer.t=t;tracer.head=head;
    return head;
  };
  tracer.setParameter(0);
  return tracer;
}

// Brief pauses let us read exact, meaningful coordinates on the heart.
const DRAW_KEYS=[[12,0],[13,0],[16,Math.PI/2],[17,Math.PI/2],[20,Math.PI],
  [21,Math.PI],[24,3*Math.PI/2],[25,3*Math.PI/2],[27,TAU],[28,TAU]];
function traceParameter(time) {
  if(time<=12)return 0;
  for(let i=1;i<DRAW_KEYS.length;i++) {
    const [end,b]=DRAW_KEYS[i],[start,a]=DRAW_KEYS[i-1];
    if(time<=end)return a+(b-a)*smooth((time-start)/(end-start));
  }
  return TAU;
}

// ── A quiet sheet of mathematical paper, not a dashboard.
const backdrop=rect({width:1280,height:720,fill:P.paper,stroke:CLEAR});
const masthead=group(label("THE QUIET SCIENCE",-471,-319,13,P.faint),label("02 / HEART",507,-319,13,P.rose));
const footer=group(label("A CURVE CALLED LOVE",-471,331,12,P.faint),label("参数 · 轨迹 · 对称",467,331,12,P.faint));

// ── 00–05 / A heart, before we see its formula.
const coverScale=fitScale(330,300),coverHeart=heartArt(coverScale,.48);
centerArt(coverHeart.g,[-267,0],coverScale);
const cover=group(coverHeart.g,label("把一个心形",258,-58,46),label("写成函数。",258,5,46),
  segment([115,54],[169,54],P.rose,2),label("让数学，慢慢画出浪漫。",279,95,21,P.muted));

// ── 05–28 / A coordinate grid, the two formulas, and an exact moving point.
const plotScale=8.9,plotOrigin=[-295,2];
const gridParts=[];
for(const x of [-16,-8,0,8,16]) {
  gridParts.push(segment([x*plotScale,-16*plotScale],[x*plotScale,20*plotScale],P.grid,.8));
  if(x!==0)gridParts.push(label(String(x),x*plotScale,18,13,P.faint));
}
for(const y of [-16,-8,0,8,16]) {
  gridParts.push(segment([-20*plotScale,-y*plotScale],[20*plotScale,-y*plotScale],P.grid,.8));
  if(y!==0)gridParts.push(label(String(y),-18,-y*plotScale,13,P.faint));
}
gridParts.push(segment([-20*plotScale,0],[20*plotScale,0],"#bcb0b4",1.2));
gridParts.push(segment([0,20*plotScale],[0,-18*plotScale],"#bcb0b4",1.2));
gridParts.push(polygon({points:[[20*plotScale,0],[20*plotScale-7,-3],[20*plotScale-7,3]],fill:P.faint,stroke:CLEAR}));
gridParts.push(polygon({points:[[0,-18*plotScale],[-3,-18*plotScale+7],[3,-18*plotScale+7]],fill:P.faint,stroke:CLEAR}));
gridParts.push(label("x",20*plotScale+15,0,16,P.muted),label("y",0,-18*plotScale-17,16,P.muted),label("0",-13,15,13,P.faint));
const reference=curvePath(POINTS.map(p=>[p[0]*plotScale,-p[1]*plotScale]),P.rose,1.2);
const tracer=makeTracer(plotScale);
const drawnFill=curvePath(POINTS.map(p=>[p[0]*plotScale,-p[1]*plotScale]),CLEAR,0,P.blush,true);
const projectionX=segment([0,0],[1,0],P.gold,1),projectionY=segment([0,0],[1,0],P.gold,1);
const xMarker=dot(0,0,2.5,P.gold),yMarker=dot(0,0,2.5,P.gold);
const pointGlow=dot(0,0,12,P.rose);pointGlow.opacity=.09;
const point=dot(0,0,4.5,P.rose);
const plotting=group(...gridParts,drawnFill,reference,tracer.g,projectionX,projectionY,xMarker,yMarker,pointGlow,point);
place(plotting,plotOrigin);

const formulaHeading=group(label("用一个参数，控制两个坐标",0,-247,39),label("t 负责前进，x(t) 与 y(t) 决定点的位置。",0,-194,22,P.muted));
const traceHeading=group(label("一个点，就能描出整颗心",0,-247,39),label("横、纵坐标同时变化，轨迹便慢慢显现。",0,-194,22,P.muted));
const formulas=group(label("x(t) = 16 sin³(t)",279,-83,26,P.rose),
  label("y(t) = 13 cos(t) − 5 cos(2t)",302,-28,22),
  label("       − 2 cos(3t) − cos(4t)",302,9,22),
  label("t : 0 → 2π",279,205,18,P.muted));

const keyPoints=[
  {t:0,name:"t = 0",value:"P = (0, 5)"},
  {t:Math.PI/2,name:"t = π / 2",value:"P = (16, 4)"},
  {t:Math.PI,name:"t = π",value:"P = (0, −17)"},
  {t:3*Math.PI/2,name:"t = 3π / 2",value:"P = (−16, 4)"},
  {t:TAU,name:"t = 2π",value:"P = (0, 5) · 回到起点"},
];
const readouts=keyPoints.map(key=>group(label(key.name,279,75,21,P.rose),label(key.value,279,111,19,P.muted)));
const movingReadout=group(label("同一个 t，同时代入两个式子",279,79,19,P.muted),label("P(t) = (x(t), y(t))",279,113,19,P.muted));
const dial=dot(279,160,25,CLEAR,P.line,1),dialPoint=dot(279,135,3,P.rose);
const coordinateLesson=group(plotting,formulaHeading,traceHeading,formulas,...readouts,movingReadout,dial,dialPoint);

// ── 28–36 / Symmetry is an identity, not a drawing trick.
const symmetryScale=fitScale(290,270),symmetryHeart=heartArt(symmetryScale,.15);
const symmetryOrigin=[-295,8];centerArt(symmetryHeart.g,symmetryOrigin,symmetryScale);
const axisX=symmetryHeart.g.x;
const axis=segment([axisX,-144],[axisX,166],P.line,1);
const mirrorA=dot(0,0,4.5,P.rose),mirrorB=dot(0,0,4.5,P.gold);
const mirrorLink=segment([0,0],[1,0],P.line,1);
const symmetry=group(label("对称，藏在三角函数里",0,-247,39),
  label("当参数换成 2π − t，左右两点恰好互为镜像。",0,-194,22,P.muted),
  axis,symmetryHeart.g,mirrorLink,mirrorA,mirrorB,
  label("x(2π − t) = −x(t)",282,-35,27,P.rose),
  label("y(2π − t) =  y(t)",282,22,27),
  label("横坐标相反，纵坐标相同。",282,97,21,P.muted));

// ── 36–44 / A closed curve, a restrained breath, a quiet ending.
const finalScale=fitScale(295,270),finalHeart=heartArt(finalScale,.5);
centerArt(finalHeart.g,[0,2],finalScale);
const finalPoint=dot(0,0,3.5,P.rose);
const ending=group(label("从公式，到心动。",0,-232,40),label("两个函数 · 一个参数 · 一条闭合曲线",0,-174,19,P.muted),
  finalHeart.g,finalPoint,label("浪漫可以被描绘，也可以被计算。",0,210,25));

const SUBTITLES=[
  {start:0,end:4.8,text:"把一个心形，写成函数。",visible:false},
  {start:5,end:11.8,text:"让横坐标和纵坐标，共同依赖参数 t。"},
  {start:12,end:15.8,text:"t 从 0 出发，点沿着右侧慢慢前进。"},
  {start:16,end:19.8,text:"同一个 t，同时决定这个点的横、纵坐标。"},
  {start:20,end:23.8,text:"经过下方的尖点，轨迹开始沿左侧返回。"},
  {start:24,end:27.8,text:"当 t 走到 2π，终点与起点重合。"},
  {start:28,end:31.8,text:"左右两点的横坐标相反，纵坐标相同。"},
  {start:32,end:35.8,text:"所以，这条心形曲线关于 y 轴对称。"},
  {start:36,end:39.8,text:"一个心形，是两个函数共同走过的轨迹。"},
  {start:40,end:43.8,text:"让公式，慢慢画出一颗心。"},
];
const captions=SUBTITLES.map(cue=>label(cue.text,0,277,23,P.muted));

const film=group(backdrop,masthead,footer,cover,coordinateLesson,symmetry,ending,...captions);
film.update((self,time)=>{
  cover.opacity=fade(time,0,5,.8);
  coordinateLesson.opacity=fade(time,5,28);
  symmetry.opacity=fade(time,28,36);
  ending.opacity=fade(time,36,44,.8);
  // Breathing changes the geometry only, never the text size.
  const coverBreath=1+.015*Math.sin(time*1.15);
  coverHeart.g.scaleX=coverBreath;coverHeart.g.scaleY=coverBreath;
  formulaHeading.opacity=1-smooth((time-11.5)/.5);
  traceHeading.opacity=smooth((time-11.5)/.5);
  const t=time<12?TAU*clamp((time-5)/7):traceParameter(time);
  const head=screenPoint(t,plotScale);
  tracer.setParameter(t);
  tracer.g.opacity=time<12?0:1;
  reference.opacity=time<12?.29:.09;
  drawnFill.opacity=time<12?0:.4*smooth((time-26)/1.2);
  place(point,head);place(pointGlow,head);
  place(xMarker,[head[0],0]);place(yMarker,[0,head[1]]);
  connect(projectionX,[head[0],0],head);connect(projectionY,[0,head[1]],head);
  // Exact coordinates are shown only during their deliberate pauses.
  let keyIndex=-1;
  if(time>=12)keyIndex=keyPoints.findIndex(key=>Math.abs(key.t-t)<1e-8);
  readouts.forEach((item,i)=>{item.opacity=i===keyIndex?1:0;});
  movingReadout.opacity=keyIndex<0?1:0;
  positionDial(t);
  const u=.15*Math.PI+.7*Math.PI*clamp((time-28)/8);
  const right=screenPoint(u,symmetryScale),left=screenPoint(TAU-u,symmetryScale);
  const a=[symmetryHeart.g.x+right[0],symmetryHeart.g.y+right[1]];
  const b=[symmetryHeart.g.x+left[0],symmetryHeart.g.y+left[1]];
  place(mirrorA,a);place(mirrorB,b);connect(mirrorLink,a,b);
  const breath=1+.022*Math.sin((time-36)*1.6);
  finalHeart.g.scaleX=breath;finalHeart.g.scaleY=breath;
  const closing=screenPoint(TAU*clamp((time-36.5)/5.4),finalScale);
  place(finalPoint,[finalHeart.g.x+closing[0]*breath,finalHeart.g.y+closing[1]*breath]);
  finalPoint.opacity=1-smooth((time-42)/.6);
  captions.forEach((item,i)=>{const cue=SUBTITLES[i];item.opacity=cue.visible===false?0:fade(time,cue.start,cue.end,.28);});
});
function positionDial(t) {place(dialPoint,[279+25*Math.sin(t),160-25*Math.cos(t)]);}
add(film);
wait(FILM.duration);
