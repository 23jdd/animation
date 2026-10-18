/// <reference path="../../animation-api.js" />

// 《把温柔，写成一朵玫瑰》 / 64 seconds / an animated botanical love letter.
// JS 绘制所有曲线、花瓣与动作；Go 引擎负责渲染。没有图片素材。
// 预览：go run . preview examples/rose-curve/animation.js
// GIF：go run . gif -fps 12 -scale 0.75 examples/rose-curve/animation.js examples/rose-curve/preview.gif
// MP4：go run . mp4 -fps 30 examples/rose-curve/animation.js silent.mp4
// 最终成片另合入 music.js 合成的轻柔配乐；所有数学角度均使用弧度。

const FILM={duration:64,n:5,samples:1600,drawStart:14.5,petalSeconds:2.2,drawSeconds:1.85};
const TAU=2*Math.PI;
const CLEAR="#00000000";
const P={paper:"#fff6f2",ink:"#664b58",rose:"#c86e8b",muted:"#a38591",faint:"#bba2aa",
  blush:"#f3ceda",petal:"#eab3c6",deep:"#bd6988",sage:"#91a28e",leaf:"#bdcbb2",gold:"#bda477"};

function label(content,x,y,size=23,fill=P.ink){return text(content,{position:[Math.round(x),Math.round(y)],size,fill,stroke:CLEAR});}
function dot(x,y,radius,fill,stroke=CLEAR,strokeWidth=0){return circle({position:[x,y],radius,fill,stroke,strokeWidth});}
function lineFrom(a,b,stroke=P.rose,strokeWidth=2){return line({start:a,end:b,stroke,strokeWidth});}
function place(object,p){object.x=p[0];object.y=p[1];}
function scale(object,s){object.scaleX=s;object.scaleY=s;}
function clamp(x){return Math.max(0,Math.min(1,x));}
function smooth(x){x=clamp(x);return x*x*(3-2*x);}
function mix(a,b,t){return a+(b-a)*t;}
function fade(time,start,end,seconds=.75){return (start===0?1:smooth((time-start)/seconds))*(1-smooth((time-end+seconds)/seconds));}
function connect(object,a,b){const x=b[0]-a[0],y=b[1]-a[1],length=Math.hypot(x,y);place(object,a);object.rotation=Math.atan2(y,x);object.scaleX=length;object.opacity=length>1e-8?1:0;}

/** Mathematical rose curve, before any decorative petal transformations.
 * r = a cos(nθ), x = r cos θ, y = r sin θ.
 * @param {number} n @param {number} theta @param {number} [a]
 * @returns {Vec2}
 */
function rose(n,theta,a=1){const r=a*Math.cos(n*theta);return [r*Math.cos(theta),r*Math.sin(theta)];}
function petalCount(n){return n%2===0?2*n:n;}
function period(n){return n%2===0?TAU:Math.PI;}
function startAngle(n){return -Math.PI/(2*n);}
function screenRose(n,theta,a){const p=rose(n,theta,a);return [p[0],-p[1]];}
function curve(points,stroke=P.rose,width=1.3,fill=CLEAR,close=false){
  const commands=points.map((p,i)=>[i===0?"M":"L",p[0],p[1]]);
  if(close)commands.push(["Z"]);
  return path({commands,stroke,strokeWidth:width,fill});
}
function ellipse(rx,ry,stroke=P.gold,width=.8,phase=0,span=TAU){
  return curve(Array.from({length:161},(_,i)=>{const a=phase+span*i/160;return [rx*Math.cos(a),ry*Math.sin(a)];}),stroke,width);
}
function heartMark(size,fill=P.rose){
  const p=path({commands:[["M",0,7],["C",-15,-2,-12,-13,-5,-11],["Q",-1,-11,0,-6],
    ["Q",4,-14,10,-9],["C",18,-1,6,5,0,7],["Z"]],fill,stroke:CLEAR});
  scale(p,size/14);return p;
}
function leaf(x,y,rotation,size=1){
  const p=path({commands:[["M",0,0],["C",10,-28,45,-30,57,-13],["C",41,4,14,11,0,0],["Z"]],
    fill:P.leaf,stroke:P.sage,strokeWidth:.8});
  const vein=path({commands:[["M",3,-1],["Q",24,-11,51,-13]],fill:CLEAR,stroke:P.sage,strokeWidth:.65});
  const g=group(p,vein);place(g,[x,y]);g.rotation=rotation;scale(g,size);return g;
}

// A decorative rose is an assembly, not an unmodified single polar curve:
// sample one true rose petal, gently widen it, then rotate and layer its copies.
// The single-line scene below shows the exact, unmodified mathematical rose.
function decorativePetal(radius,widen,fill){
  const points=[];
  for(let i=0;i<=112;i++){
    const theta=-Math.PI/10+Math.PI/5*i/112,p=rose(5,theta,radius);
    points.push([p[0],-p[1]*widen]);
  }
  const body=curve(points,"#d38fa5",.8,fill,true);
  const fold=curve(points.slice(23,91).map(([x,y])=>[x*.88,y*.78]),"#fff2ee",.85);
  fold.opacity=.58;
  return group(body,fold);
}
function flower(radius,tint=0){
  const rings=[],parts=[];
  const ratios=[1,.82,.64,.46,.29];
  const colors=tint===1?["#f4d7e1","#ecc2d4","#e0adc4","#dca6c0","#cf92b1"]:
    ["#f8dce3","#f3ceda","#edbdcf","#e5abc2","#d996b3"];
  ratios.forEach((ratio,k)=>{
    const petals=[];
    for(let j=0;j<5;j++){
      const p=decorativePetal(radius*ratio,2.65-k*.12,colors[k]);
      p.rotation=-Math.PI/2-j*TAU/5+k*.49;
      petals.push(p);
    }
    const g=group(...petals);parts.push(g);rings.push({g,petals});
  });
  const curl=[];
  for(let i=0;i<=160;i++){
    const t=3.5*Math.PI*i/160,r=radius*.095*Math.exp(-.18*t);
    curl.push([r*Math.cos(t),-r*Math.sin(t)]);
  }
  const centre=group(curve(curl,"#bb7290",1.1),dot(0,0,1.8,"#fbe2d9"));
  parts.push(centre);
  const dew=dot(radius*.27,-radius*.55,2.2,"#fffaf2");dew.opacity=.75;parts.push(dew);
  return {g:group(...parts),rings,centre};
}

function stemmedFlower(radius,tint=0){
  const stem=path({commands:[["M",0,60],["C",-10,110,15,164,0,223]],fill:CLEAR,stroke:P.sage,strokeWidth:2});
  const bloom=flower(radius,tint);
  const g=group(stem,leaf(-1,153,-.38,.9),leaf(4,186,Math.PI+.2,.75),bloom.g);
  return {g,bloom};
}

// Exact parameter-based tracing, batched into 50 paths plus 32 reusable lines.
// It does not confuse parameter speed with distance travelled along the curve.
function makeTracer(n,radius,ink=P.rose){
  const count=FILM.samples,chunk=32,t0=startAngle(n),span=period(n);
  const points=Array.from({length:count+1},(_,i)=>screenRose(n,t0+span*i/count,radius));
  const blocks=[];
  for(let i=0;i<count;i+=chunk)blocks.push(curve(points.slice(i,i+chunk+1),ink,2.1));
  const partial=Array.from({length:chunk},()=>lineFrom([0,0],[1,0],ink,2.1));
  const fills=[];
  for(let j=0;j<petalCount(n);j++){
    const points=Array.from({length:161},(_,i)=>screenRose(n,t0+(j+i/160)*Math.PI/n,radius));
    fills.push(curve(points,CLEAR,0,["#f0c5d4","#edd3e1","#f5d7db","#e8c6d8","#f4ccd6"][j%5],true));
  }
  const tracer={g:group(...fills,...blocks,...partial),blocks,partial,fills,head:[0,0],theta:t0,progress:0};
  tracer.setProgress=progress=>{
    progress=clamp(progress);const theta=t0+span*progress;
    const sample=Math.min(count,Math.floor(progress*count)),block=Math.floor(sample/chunk),first=block*chunk;
    blocks.forEach((p,i)=>{p.opacity=i<block?1:0;});
    const head=screenRose(n,theta,radius);
    partial.forEach((line,j)=>{const i=first+j;if(i>=count||i>sample){line.opacity=0;return;}connect(line,points[i],i===sample?head:points[i+1]);});
    fills.forEach((p,j)=>{p.opacity=.5*smooth((progress*petalCount(n)-j-.88)/.12);});
    tracer.head=head;tracer.theta=theta;tracer.progress=progress;return head;
  };
  tracer.setProgress(0);return tracer;
}
function drawingProgress(time){
  const elapsed=Math.max(0,time-FILM.drawStart),petal=Math.floor(elapsed/FILM.petalSeconds);
  if(petal>=FILM.n)return 1;
  return (petal+smooth((elapsed-petal*FILM.petalSeconds)/FILM.drawSeconds))/FILM.n;
}

// Airy botanical detail, built once. The root updater only changes transforms.
function sprig(height,seed=0){
  const parts=[path({commands:[["M",0,0],["Q",9,-height*.54,0,-height]],fill:CLEAR,stroke:P.sage,strokeWidth:.85})];
  for(let j=0;j<6;j++)for(const side of [-1,1]){
    const y=-height*(.18+j*.13),x=side*(14+14*Math.sin(j*.8+seed)**2),tipY=y-height*.13;
    parts.push(path({commands:[["M",3,y],["Q",x*.55,y-5,x,tipY]],fill:CLEAR,stroke:P.sage,strokeWidth:.65}));
    parts.push(leaf(x*.6,y-6,side<0?Math.PI+.4:-.8,.2));
    for(let k=0;k<4;k++){
      const a=k*TAU/4;parts.push(dot(x+3*Math.cos(a),tipY+3*Math.sin(a),2.65,"#fffdf5","#dbcdb6",.35));
    }
    parts.push(dot(x,tipY,1.05,P.gold));
  }
  return group(...parts);
}
function bud(){
  return group(
    path({commands:[["M",0,21],["C",-20,9,-16,-15,-5,-16],["Q",1,-23,8,-13],["C",21,0,14,12,0,21],["Z"]],fill:"#eab9cb",stroke:"#cc8ca4",strokeWidth:.9}),
    path({commands:[["M",-8,-12],["Q",8,0,0,17],["M",8,-11],["Q",-5,0,0,17]],fill:CLEAR,stroke:"#f8e1e4",strokeWidth:1}),
    path({commands:[["M",0,23],["Q",-14,22,-17,9],["Q",-8,17,0,18],["Q",8,17,15,8],["Q",13,21,0,23],["Z"]],fill:P.leaf,stroke:P.sage,strokeWidth:.7}));
}
function petalShape(fill){
  return path({commands:[["M",0,-9],["C",-10,-5,-10,6,0,10],["C",8,7,10,-5,0,-9],["Z"]],fill,stroke:CLEAR});
}
function butterfly(){
  const left=group(
    path({commands:[["M",0,0],["C",-24,-27,-29,-1,0,5],["Z"]],fill:"#edd4df",stroke:"#c9a2b4",strokeWidth:.6}),
    path({commands:[["M",0,2],["C",-27,1,-14,23,0,6],["Z"]],fill:"#f0dce6",stroke:"#c9a2b4",strokeWidth:.6}));
  const right=group(
    path({commands:[["M",0,0],["C",24,-27,29,-1,0,5],["Z"]],fill:"#f2d3db",stroke:"#c9a2b4",strokeWidth:.6}),
    path({commands:[["M",0,2],["C",27,1,14,23,0,6],["Z"]],fill:"#f4dfe3",stroke:"#c9a2b4",strokeWidth:.6}));
  return {g:group(left,right,lineFrom([0,-3],[0,10],P.muted,1.2)),left,right};
}

const backdrop=rect({width:1280,height:720,fill:P.paper,stroke:CLEAR});
const masthead=group(label("A LITTLE ROMANCE",-480,-319,13,P.faint),label("03 / IN BLOOM",491,-319,13,P.rose));
const footer=group(label("DRAWN WITH LOVE, AND A LITTLE MATH",-414,332,11,P.faint),label("温柔 · 绽放 · 心意",467,332,12,P.faint));

const floating=Array.from({length:18},(_,i)=>{
  const side=i%2?-1:1,x=side*(430+(i*37)%128),y=-259+(i*79)%520;
  const p=petalShape(i%3?"#f0ccd9":"#e4bdd0");scale(p,.35+(i%4)*.14);
  return {p,x,y,i};
});
const sparkles=Array.from({length:24},(_,i)=>{
  const p=group(lineFrom([-3,0],[3,0],P.gold,.8),lineFrom([0,-3],[0,3],P.gold,.8));
  place(p,[(i%2?-1:1)*(360+(i*41)%215),-232+(i*71)%450]);return p;
});

// I / A botanical title with a slowly opening bloom.
const coverRose=stemmedFlower(119);place(coverRose.g,[-271,-52]);
const coverArc=ellipse(166,180,P.gold,.7,-2.5,4.1);place(coverArc,[-271,-20]);coverArc.opacity=.3;
const coverSprigs=[sprig(134,1),sprig(97,2)];
place(coverSprigs[0],[-346,169]);coverSprigs[0].rotation=-.72;
place(coverSprigs[1],[-197,169]);coverSprigs[1].rotation=.66;
const cover=group(coverArc,...coverSprigs,coverRose.g,label("把温柔，",244,-70,46),label("写成一朵玫瑰。",302,-8,43),
  lineFrom([115,44],[167,44],P.rose,1.5),label("有些心意，适合慢慢盛开。",290,90,21,P.muted),label("A FLOWER, A LETTER, A LITTLE LOVE",275,155,12,P.faint));

// II / The envelope opens and a small rose emerges on the letter.
const letterRose=makeTracer(5,49);place(letterRose.g,[0,2]);letterRose.g.rotation=-Math.PI/2;
const letter=group(rect({width:230,height:177,radius:3,fill:"#fffdf8",stroke:"#e1d0c3",strokeWidth:1}),
  label("FOR YOU",0,-60,12,P.rose),letterRose.g,label("愿你被温柔以待",0,64,14,P.muted));
const envelopeBack=rect({width:288,height:187,position:[0,25],radius:5,fill:"#f1dfd9",stroke:"#d8b8bc",strokeWidth:1});
const envelopePocket=path({commands:[["M",-144,-34],["L",0,49],["L",144,-34],["L",144,118],["L",-144,118],["Z"]],fill:"#f9eae2",stroke:"#d8b8bc",strokeWidth:1});
const envelopeCrease=path({commands:[["M",-143,117],["L",0,25],["L",143,117]],fill:CLEAR,stroke:"#dec2c0",strokeWidth:1});
const flap=path({commands:[["M",-143,-34],["L",143,-34],["L",0,57],["Z"]],fill:"#eed9d4",stroke:"#d8b8bc",strokeWidth:1});
const seal=group(dot(0,0,14,"#dba8b6","#c98fa3",.7),heartMark(7,"#fff4ed"));place(seal,[0,58]);
const envelope=group(envelopeBack,letter,envelopePocket,envelopeCrease,flap,seal);place(envelope,[-260,12]);
const prelude=group(envelope,label("先寄给你，一点小小的心意。",230,-108,29),
  label("不必急着说出口，",243,-39,23,P.muted),label("让它先开成一朵花。",243,1,23,P.muted),
  label("r(θ) = a cos(5θ)",243,77,23,P.rose),label("这条曲线，有一个温柔的名字：玫瑰线。",0,251,22,P.muted));

// III / An exact mathematical curve, with a continuous, parameter-matched head.
const tracer=makeTracer(5,151);
const reference=curve(Array.from({length:1601},(_,i)=>screenRose(5,startAngle(5)+Math.PI*i/1600,151)),P.rose,.8);reference.opacity=.1;
const tracingHead=dot(0,0,3.6,P.gold),headGlow=dot(0,0,11,P.gold);headGlow.opacity=.09;
const drawingGroup=group(reference,tracer.g,headGlow,tracingHead);place(drawingGroup,[0,-25]);
const petalBadges=Array.from({length:5},(_,j)=>{
  const p=dot(-52+j*26,210,2.5,P.rose);return p;
});
const sideNotes=group(label("一笔",-311,-66,19,P.muted),label("一瓣",309,29,19,P.muted),
  label("一点心意",-311,-30,13,P.faint),label("慢慢盛开",309,65,13,P.faint));
const drawing=group(label("一笔，也能藏住一整朵温柔。",0,-248,33),drawingGroup,sideNotes,...petalBadges,
  label("r = a cos(5θ)",0,179,17,P.muted),label("五片花瓣，从一点心意里慢慢长出。",0,251,23,P.ink));

// IV / Three distinct complete rose curves, not a misleading noninteger morph.
const galleryItems=[3,5,7].map((n,i)=>{
  const trace=makeTracer(n,105,["#bd8aa5",P.rose,"#b28fba"][i]);trace.g.rotation=-Math.PI/2;
  const halo=ellipse(127,127,"#ddc8c2",.7);halo.opacity=.35;
  const g=group(halo,trace.g,label(["初见","心动","相伴"][i],0,151,25),label(`r = a cos(${n}θ)`,0,188,15,P.muted));
  place(g,[(i-1)*320,-37]);return {n,trace,g};
});
const gallery=group(label("同一份心意，也有不同的盛开。",0,-256,32),...galleryItems.map(o=>o.g),
  label("三瓣的初见，五瓣的心动，七瓣的长久相伴。",0,251,22,P.muted));

// V / The exact petal becomes an explicitly decorative, layered flower.
const openingRose=stemmedFlower(137);place(openingRose.g,[0,-48]);
const openingSprigs=[sprig(160,2),sprig(145,3),sprig(104,4),sprig(107,5)];
const sprigAnchors=[[-149,136,-.55],[148,142,.6],[-234,122,-.3],[233,118,.32]];
openingSprigs.forEach((g,i)=>{const [x,y,r]=sprigAnchors[i];place(g,[x,y]);g.rotation=r;});
const opening=group(label("把靠近，留给时间。",0,-256,33),...openingSprigs,openingRose.g);

// Inward spiral: the petals arrive in a finite, continuous motion, then vanish.
const spiralPetals=Array.from({length:30},(_,i)=>{
  const p=petalShape(["#edc2d0","#ddacc4","#f1d5d6"][i%3]);return {p,i};
});
function spiralState(i,time){
  const u=clamp((time-36-i*.025)/5.2),angle=i*2.39996+u*TAU*1.1,radius=mix(285,19,smooth(u));
  return {x:radius*Math.cos(angle),y:-35+.68*radius*Math.sin(angle),rotation:angle+.4,
    opacity:.56*Math.sin(Math.PI*u)**2,scale:.3+.35*(1-u)};
}

// VI / A twelve-flower wreath and two small butterflies.
const wreathRing=ellipse(254,166,"#b5bea6",1);wreathRing.opacity=.55;place(wreathRing,[0,-17]);
const wreathItems=Array.from({length:12},(_,i)=>{
  const theta=-Math.PI/2+i*TAU/12,petal=flower(i%3===0?29:22,i%2);
  const foliage=group(leaf(-18,7,Math.PI-.3,.5),leaf(12,11,-.8,.5));
  const g=group(foliage,petal.g);place(g,[254*Math.cos(theta),-17+166*Math.sin(theta)]);
  return {g,theta};
});
const wreath=group(label("让所有的小美好，绕着你盛开。",0,-256,32),wreathRing,...wreathItems.map(o=>o.g),
  label("你在，",0,-48,32),label("便是花期。",0,7,39,P.rose),label("不止这一刻，更是每一个平常的日子。",0,251,22,P.muted));
const butterflies=[butterfly(),butterfly()];
const butterflyLayer=group(...butterflies.map(b=>b.g));

// VII / Seven roses, buds, baby's breath, folded paper, a bow and a tiny tag.
const bouquetAnchors=[[-185,-32],[-108,-113],[0,-151],[109,-109],[191,-31],[-58,-30],[66,-18]];
const bouquetFlowers=[66,76,87,73,65,78,77].map((r,i)=>flower(r,i%3===0?1:0));
const wrap=group(
  path({commands:[["M",-224,-3],["L",-138,159],["L",8,204],["L",40,81],["Z"]],fill:"#f2ded8",stroke:"#d8b6b5",strokeWidth:1}),
  path({commands:[["M",218,-3],["L",147,151],["L",-8,204],["L",-46,80],["Z"]],fill:"#f9e9e0",stroke:"#dbbcb7",strokeWidth:1}),
  path({commands:[["M",-223,-3],["Q",-109,83,0,194],["Q",117,75,218,-3]],fill:CLEAR,stroke:"#c3aa83",strokeWidth:.8}));
const stemParts=bouquetAnchors.map(([x,y])=>path({commands:[["M",0,182],["C",x*.35,120,x*.94,25,x,y]],fill:CLEAR,stroke:P.sage,strokeWidth:1.35}));
const bouquetSprigs=[sprig(172,1),sprig(169,2),sprig(146,3),sprig(150,4)];
const bouquetSprigAnchors=[[-92,91,-.95],[89,91,.97],[-41,45,-.35],[50,47,.4]];
bouquetSprigs.forEach((g,i)=>{const [x,y,r]=bouquetSprigAnchors[i];place(g,[x,y]);g.rotation=r;});
const bouquetBuds=[bud(),bud(),bud()];
const budAnchors=[[-238,-86,-.45],[236,-71,.44],[179,-157,.29]];
bouquetBuds.forEach((g,i)=>{
  const [x,y,r]=budAnchors[i];place(g,[x,y]);g.rotation=r;
  stemParts.push(path({commands:[["M",0,160],["Q",x*.85,65,x,y+20]],fill:CLEAR,stroke:P.sage,strokeWidth:1}));
});
const stems=group(...stemParts,leaf(-72,108,Math.PI+.3,.7),leaf(75,111,-.5,.75));
const bowLoops=path({commands:[["M",0,0],["C",-66,-46,-71,31,0,0],["C",64,-43,72,29,0,0]],fill:"#edcbd2",stroke:"#c98ba0",strokeWidth:1.3});
const bowTails=path({commands:[["M",0,0],["Q",-12,33,-45,52],["M",0,0],["Q",9,37,39,51]],fill:CLEAR,stroke:"#d69bae",strokeWidth:2.3});
const bow=group(bowTails,bowLoops,dot(0,0,4,"#cd98a9"));place(bow,[0,139]);
const giftTag=group(lineFrom([0,0],[25,32],P.gold,.8),
  rect({width:42,height:57,position:[33,57],radius:3,fill:"#fff9ee",stroke:"#d8c1a6",strokeWidth:.8}),
  dot(33,37,1.5,P.gold),label("予你",33,57,12,P.muted));place(giftTag,[28,124]);
const bouquet=group(label("把这一束，送给心上的人。",0,-276,32),wrap,stems,...bouquetSprigs,...bouquetBuds,
  ...bouquetFlowers.map(f=>f.g),bow,giftTag,label("愿每一天，都有一点被偏爱的浪漫。",0,260,23,P.ink),
  label("七朵花，一封信，和想陪你很久的心意。",0,298,14,P.muted));

const CUES=[
  {start:36,end:40,text:"让花瓣，一层一层地拥抱。"},
  {start:40.2,end:44.7,text:"把未说出口的话，轻轻交给晚风。"},
];
const captions=CUES.map(c=>label(c.text,0,252,23,P.muted));
const SCENES=[{g:cover,start:0,end:7},{g:prelude,start:7,end:14},{g:drawing,start:14,end:27},
  {g:gallery,start:27,end:36},{g:opening,start:36,end:45},{g:wreath,start:45,end:54},{g:bouquet,start:54,end:64}];
const film=group(backdrop,...floating.map(f=>f.p),...sparkles,masthead,footer,...SCENES.map(s=>s.g),
  ...spiralPetals.map(o=>o.p),butterflyLayer,...captions);
film.update((self,time)=>{
  SCENES.forEach(s=>{s.g.opacity=fade(time,s.start,s.end,.85);});
  floating.forEach(({p,x,y,i})=>{p.x=x+14*Math.sin(time*.22+i);p.y=y+21*Math.sin(time*.17+i*.8);
    p.rotation=.3*Math.sin(time*.26+i)+i*.61;p.opacity=.14+.09*Math.sin(time*.18+i);});
  sparkles.forEach((p,i)=>{p.opacity=.09+.11*Math.sin(time*.66+i*1.9)**8;});
  coverRose.g.rotation=.012*Math.sin(time*.8);
  coverRose.bloom.rings.forEach((ring,k)=>{const u=smooth((time-k*.22)/1.6);ring.g.opacity=u;scale(ring.g,.82+.18*u);});
  coverSprigs.forEach((g,i)=>{g.rotation=[-.72,.66][i]+.025*Math.sin(time*.7+i);});
  const letterOpen=smooth((time-7.6)/2.1);
  letter.y=mix(15,-61,letterOpen);flap.opacity=1-smooth((time-7.5)/.7);
  seal.opacity=1-smooth((time-7.5)/.65);letterRose.setProgress(smooth((time-9)/3.3));
  envelope.rotation=.008*Math.sin(time*.6);
  const head=tracer.setProgress(drawingProgress(time));place(tracingHead,head);place(headGlow,head);
  tracingHead.opacity=1-smooth((time-25.5)/.6);headGlow.opacity=.09*tracingHead.opacity;
  petalBadges.forEach((p,j)=>{p.opacity=.18+.62*smooth((tracer.progress*5-j-.8)/.2);});
  sideNotes.opacity=.68+.15*Math.sin(time*.7);
  galleryItems.forEach((item,i)=>{
    const u=smooth((time-27.6-i*.6)/4.4);item.trace.setProgress(u);
    scale(item.trace.g,.94+.06*smooth((time-27-i*.6)/3));
    item.g.y=-37+3*Math.sin(time*.7+i);item.g.opacity=smooth((time-27-i*.45)/.9);
  });
  openingRose.bloom.rings.forEach((ring,k)=>{
    const u=smooth((time-36-k*.5)/2.7);scale(ring.g,.38+.62*u);ring.g.opacity=u;
    ring.g.rotation=(1-u)*(.23+k*.025)+.012*Math.sin(time*.7+k);
  });
  openingRose.bloom.centre.opacity=smooth((time-38.7)/1.6);openingRose.g.rotation=.009*Math.sin(time*.65);
  openingSprigs.forEach((g,i)=>{g.opacity=smooth((time-37.6-i*.3)/2.1);g.rotation=sprigAnchors[i][2]+.025*Math.sin(time*.7+i);});
  spiralPetals.forEach(({p,i})=>{const s=spiralState(i,time);place(p,[s.x,s.y]);p.rotation=s.rotation;p.opacity=s.opacity;scale(p,s.scale);});
  wreathItems.forEach((item,i)=>{const u=smooth((time-45-i*.2)/1.7);item.g.opacity=u;scale(item.g,.48+.52*u);item.g.rotation=.03*Math.sin(time*.6+i);});
  wreathRing.opacity=.55*smooth((time-45)/3.5);
  butterflies.forEach((b,i)=>{
    const t=(time-45)*.6+i*2.6;
    place(b.g,[(i?1:-1)*(299+24*Math.sin(t)), -28+127*Math.sin(t*.77+i)]);
    b.g.rotation=(i?-.16:.16)+.15*Math.sin(t);scale(b.g,i?.43:.52);
    b.left.scaleX=b.right.scaleX=.35+.65*(.5+.5*Math.cos(time*7+i));
    b.g.opacity=fade(time,46+i*.4,54,.7);
  });
  wrap.opacity=smooth((time-54.5)/1.6);wrap.scaleY=.8+.2*smooth((time-54.5)/1.6);
  stems.opacity=smooth((time-54)/1.4);
  bouquetSprigs.forEach((g,i)=>{g.opacity=smooth((time-54.4-i*.15)/1.8);g.rotation=bouquetSprigAnchors[i][2]+.016*Math.sin(time*.8+i);});
  bouquetBuds.forEach((g,i)=>{g.opacity=smooth((time-55-i*.22)/1.4);g.rotation=budAnchors[i][2]+.017*Math.sin(time*.9+i);});
  bouquetFlowers.forEach((flower,i)=>{
    const grow=smooth((time-54.2-i*.21)/2.1),breath=1+.007*Math.sin(time*.8+i),anchor=bouquetAnchors[i];
    scale(flower.g,(.35+.65*grow)*breath);place(flower.g,[anchor[0]*(1.24-.24*grow),anchor[1]+74*(1-grow)]);
    flower.g.rotation=(i%2?-.2:.2)*(1-grow)+.009*Math.sin(time*.69+i);flower.g.opacity=grow;
  });
  const tie=smooth((time-56.5)/1.3);scale(bowLoops,.12+.88*tie);bow.opacity=tie;
  bowTails.rotation=.025*Math.sin(time*.8);giftTag.opacity=smooth((time-57.3)/1);giftTag.rotation=.025*Math.sin(time*.9);
  captions.forEach((item,i)=>{const cue=CUES[i];item.opacity=fade(time,cue.start,cue.end,.32);});
});
add(film);
wait(FILM.duration);
