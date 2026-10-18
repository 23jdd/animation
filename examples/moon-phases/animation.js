/// <reference path="../../animation-api.js" />

// 《月亮，为什么会变？》 / A quiet science film / 44 seconds.
// 全部图形和动画由当前引擎的 JS API 绘制，无图片素材，无外部 JS 依赖。
// 预览：go run . preview examples/moon-phases/animation.js
// 视频：go run . mp4 -fps 30 examples/moon-phases/animation.js examples/moon-phases/moon-phases.mp4
// GIF：go run . gif -fps 12 -scale 0.75 examples/moon-phases/animation.js examples/moon-phases/preview.gif
// 画布：1280×720。文字使用原生字号，不缩放文字，避免静止时发虚。
// 科学参考：https://science.nasa.gov/moon/moon-phases/
//           https://science.nasa.gov/moon/eclipses/

const FILM = { duration: 44, orbitStart: 12, orbitEnd: 28 };
const TAU = Math.PI * 2;
const CLEAR = "#00000000";
const P = {
  night: "#0c1726", ink: "#142338", line: "#344557", moon: "#eee6d2",
  shadow: "#29374b", white: "#f0ece1", muted: "#a5b1bd", faint: "#748698",
  gold: "#d9bc83", blue: "#719fae", sea: "#35566c", mint: "#a5c3b4",
};

function textAt(content, x, y, size = 22, fill = P.white) {
  return text(content, {position:[Math.round(x),Math.round(y)],size,fill,stroke:CLEAR});
}
function dot(x, y, radius, fill, stroke = CLEAR, strokeWidth = 0) {
  return circle({position:[x,y],radius,fill,stroke,strokeWidth});
}
function lineBetween(a, b, stroke = P.line, strokeWidth = 1) {
  return line({start:a,end:b,stroke,strokeWidth});
}
function position(object, x, y) { object.x=x;object.y=y; }
function clamp(x) { return Math.max(0,Math.min(1,x)); }
function smooth(x) { x=clamp(x);return x*x*(3-2*x); }
function windowOpacity(time, start, end, fade = 0.65) {
  return (start===0 ? 1 : smooth((time-start)/fade)) * (1-smooth((time-end+fade)/fade));
}

/** New Moon = 0; first quarter = PI/2; full Moon = PI; last quarter = 3PI/2.
 * The brightness is the projected sunlit area, not an occluding Earth shadow.
 * @param {number} angle
 */
function phaseState(angle) {
  const a=((angle%TAU)+TAU)%TAU;
  return { waxing:a<=Math.PI, cosine:Math.cos(a), fraction:(1-Math.cos(a))/2 };
}

// Pause briefly at the four key phases so their names can be read clearly.
// Angle is measured relative to the Sun direction, starting at new Moon.
function orbitalAngle(time) {
  const keys=[[12,0],[13,0],[16,Math.PI/2],[17,Math.PI/2],[20,Math.PI],
    [21,Math.PI],[24,3*Math.PI/2],[25,3*Math.PI/2],[27,TAU],[28,TAU]];
  if(time<=12)return 0;
  for(let i=1;i<keys.length;i++) {
    if(time<=keys[i][0]) {
      const [start,a]=keys[i-1], [end,b]=keys[i];
      return a+(b-a)*smooth((time-start)/(end-start));
    }
  }
  return TAU;
}

function phaseNameIndex(angle) {
  const epsilon=1e-7;
  if(angle<epsilon||TAU-angle<epsilon)return 0;
  if(Math.abs(angle-Math.PI/2)<epsilon)return 2;
  if(Math.abs(angle-Math.PI)<epsilon)return 4;
  if(Math.abs(angle-3*Math.PI/2)<epsilon)return 6;
  return angle<Math.PI/2?1:angle<Math.PI?3:angle<3*Math.PI/2?5:7;
}

/** Dense half-disc contour; curved lunar limbs are not a few straight lines. */
function halfDisc(radius, right, fill) {
  const commands=[];
  const start=right ? -Math.PI/2 : Math.PI/2;
  for(let i=0;i<=96;i++) {
    const a=start+Math.PI*i/96;
    commands.push([i===0?"M":"L",radius*Math.cos(a),radius*Math.sin(a)]);
  }
  commands.push(["Z"]);
  return path({commands,fill,stroke:CLEAR,strokeWidth:0});
}

// A phase is a lit half-disc plus/minus a projected terminator ellipse.
// Its horizontal radius is |cos(angle)| * R: correct crescent/quarter/gibbous shapes.
// Only geometric ellipses are scaled; no bitmap or text is stretched.
function moon(radius, craters = true) {
  const waxing=halfDisc(radius,true,P.moon);
  const waning=halfDisc(radius,false,P.moon);
  const litEllipse=dot(0,0,radius,P.moon);
  const darkEllipse=dot(0,0,radius,P.shadow);
  const surface=[];
  if(craters) {
    for(const [x,y,r] of [[-.43,-.21,.14],[.26,.31,.17],[.17,-.46,.09],[-.16,.49,.08],[.55,-.17,.06],[-.48,.32,.065],[.08,.04,.055]]) {
      const crater=dot(x*radius,y*radius,r*radius,"#26364a");
      crater.opacity=0.09;surface.push(crater);
    }
  }
  const outline=dot(0,0,radius,CLEAR,"#95a1ac",0.7);outline.opacity=0.24;
  const g=group(dot(0,0,radius,P.shadow),waxing,waning,litEllipse,darkEllipse,...surface,outline);
  function phase(angle) {
    const state=phaseState(angle), c=state.cosine;
    waxing.opacity=state.waxing ? 1 : 0;waning.opacity=state.waxing ? 0 : 1;
    litEllipse.opacity=c<0 ? 1 : 0;darkEllipse.opacity=c>=0 ? 1 : 0;
    litEllipse.scaleX=Math.abs(c);darkEllipse.scaleX=Math.abs(c);
    return state;
  }
  phase(0);
  return {g,phase};
}

function sun(radius) {
  const rings=[1.65,1.35,1.12].map((factor,i)=>{
    const glow=dot(0,0,radius*factor,P.gold);glow.opacity=0.025+i*0.014;return glow;
  });
  return group(...rings,dot(0,0,radius,"#e3c797"),dot(-radius*.16,-radius*.16,radius*.66,"#ead6b0"));
}

function earth(radius) {
  const light=halfDisc(radius,false,P.blue);
  const land=path({commands:[
    ["M",-radius*.65,-radius*.4],["Q",-radius*.13,-radius*.75,radius*.06,-radius*.26],
    ["Q",-radius*.5,-radius*.08,-radius*.19,radius*.3],
    ["Q",-radius*.37,radius*.68,-radius*.57,radius*.25],["Z"],
  ],fill:P.mint,stroke:CLEAR});land.opacity=.65;
  return group(dot(0,0,radius+4,CLEAR,"#42657a",1),dot(0,0,radius,P.sea),light,land);
}

// ── Art direction: quiet navy, warm moonlight, generous space, no dashboard UI.
const backdrop=rect({width:1280,height:720,fill:P.night,stroke:CLEAR});
const stars=[];
for(let i=0;i<54;i++) {
  const x=-604+((i*317+83)%1208), y=-306+((i*193+31)%614);
  const star=dot(x,y,i%9===0 ? 1.2 : .65,P.white);
  star.opacity=.1;stars.push(star);
}
const masthead=group(textAt("THE QUIET SCIENCE",-471,-319,13,P.faint),textAt("01 / MOON",512,-319,13,P.gold));
const footer=group(textAt("月相 / PHASES OF THE MOON",-436,331,12,P.faint),textAt("光 · 轨道 · 视角",478,331,12,P.faint));

// ── 01 / The question, 0–5 s.
const coverMoon=moon(118);position(coverMoon.g,-254,-4);
const coverHalo=dot(-254,-4,154,CLEAR,P.line,.8);coverHalo.opacity=.4;
const cover=group(coverHalo,coverMoon.g,
  textAt("月亮，",225,-69,52),textAt("为什么会变？",303,-1,52),
  lineBetween([104,52],[158,52],P.gold,2),
  textAt("关于光与视角的 44 秒",286,91,21,P.muted));

// ── 02 / Moonlight is reflected sunlight, 5–12 s.
const lamp=sun(47);position(lamp,-358,18);
const litMoon=moon(93);position(litMoon.g,240,18);litMoon.phase(3*Math.PI/2);
const beams=[],photons=[];
for(let i=0;i<3;i++) {
  const y=-18+i*36;
  const beam=lineBetween([-274,y],[143,y],"#6c6454",1);beam.opacity=.56;beams.push(beam);
  const head=polygon({points:[[143,y],[135,y-4],[135,y+4]],fill:P.gold,stroke:CLEAR});head.opacity=.5;beams.push(head);
  const photon=dot(-240,y,2.2,P.gold);photons.push(photon);
}
const light=group(textAt("月亮不自己发光",0,-239,42),
  textAt("我们看见的月光，来自太阳的反射。",0,-185,23,P.muted),
  lamp,...beams,...photons,litMoon.g,textAt("太阳",-358,125,22,P.gold),
  textAt("亮面",190,158,18,P.moon),textAt("暗面",291,158,18,P.muted));

// ── 03 / Link an orbit view to the view from Earth, 12–28 s.
const orbitCenter=[-257,0], orbitRadius=146;
const orbitalSun=sun(32);position(orbitalSun,-565,0);
const planet=earth(30);position(planet,...orbitCenter);
const orbitLine=dot(...orbitCenter,orbitRadius,CLEAR,P.line,1.2);
const orbitMarks=[];
for(let i=0;i<48;i++) {
  const angle=i*TAU/48;
  const mark=dot(orbitCenter[0]+orbitRadius*Math.cos(angle),orbitRadius*Math.sin(angle),i%12===0?2.1:.8,P.muted);
  mark.opacity=i%12===0?.42:.14;orbitMarks.push(mark);
}
const smallMoon=moon(20,false);smallMoon.phase(3*Math.PI/2);
const movingMoon=group(dot(0,0,27,CLEAR,"#496079",.8),smallMoon.g,textAt("月球",0,39,16,P.muted));
const earthView=moon(104);position(earthView.g,350,0);
const phaseNames=["新月","蛾眉月","上弦月","盈凸月","满月","亏凸月","下弦月","残月"];
const phaseLabels=phaseNames.map(name=>textAt(name,350,154,27));
const orbital=group(textAt("改变的是，我们看见的角度",0,-248,39),
  textAt("同一轮月亮，在不同位置呈现不同月相。",0,-195,22,P.muted),
  orbitalSun,textAt("太阳",-565,66,17,P.gold),orbitLine,...orbitMarks,
  lineBetween([-513,0],[-447,0],P.gold,1.2),
  polygon({points:[[-444,0],[-452,-4],[-452,4]],fill:P.gold,stroke:CLEAR}),
  planet,textAt("地球",-257,58,17,P.muted),movingMoon,
  lineBetween([54,-146],[54,176],"#25374a",1),
  textAt("从地球上看到的月相",350,-149,18,P.gold),earthView.g,...phaseLabels,
  textAt("轨道与大小均为示意；倾角已省略",-257,215,13,P.faint),
  textAt("亮面朝向采用北半球常见示意",350,215,13,P.faint));

// ── 04 / Phases are not lunar eclipses, 28–35 s.
const phaseExample=moon(73);position(phaseExample.g,-305,4);
const eclipseEarth=earth(35);position(eclipseEarth,145,4);
const shadowCone=polygon({points:[[180,-31],[493,-15],[493,23],[180,39]],fill:"#07101e",stroke:CLEAR});
const eclipsedMoon=moon(23);eclipsedMoon.phase(Math.PI);position(eclipsedMoon.g,361,4);
const eclipseTint=dot(361,4,23,"#95604c");
const eclipse=group(textAt("月相，不是月食",0,-246,42),
  textAt("日常的阴晴圆缺，不是地球影子造成的。",0,-192,23,P.muted),
  lineBetween([0,-122],[0,166],"#25374a",1),
  textAt("月相",-305,-128,27,P.moon),phaseExample.g,textAt("看见的亮面比例改变",-305,149,21,P.muted),
  textAt("月食",302,-128,27,P.gold),shadowCone,eclipsedMoon.g,eclipseTint,eclipseEarth,
  textAt("地球的影子",326,-63,17,P.faint),textAt("地球",145,86,17,P.muted),textAt("月球",361,86,17,P.muted),
  textAt("进入地球的影子",302,149,21,P.muted),textAt("月食原理示意，非真实比例",302,205,13,P.faint));

// ── 05 / One lunation, 35–44 s.
const phaseStrip=[],phaseRings=[];
for(let i=0;i<8;i++) {
  const x=-490+i*140, item=moon(31,false);item.phase(i*TAU/8);position(item.g,x,39);
  const ring=dot(x,39,42,CLEAR,P.gold,1.1);phaseRings.push(ring);
  phaseStrip.push(item.g,ring,textAt(phaseNames[i],x,113,18,P.muted));
}
const cycle=group(textAt("LUNAR CYCLE",0,-230,15,P.gold),textAt("约 29.5 天",0,-150,58),
  textAt("从一次新月，到下一次新月。",0,-91,23,P.muted),
  lineBetween([-537,39],[537,39],"#25374a",1),...phaseStrip,textAt("一轮月，万千相。",0,216,31));

// Narration / burned-in subtitles. Matching SRT is supplied alongside the film.
const SUBTITLES=[
  {start:0,end:4.8,content:"月亮，为什么会有阴晴圆缺？",visible:false},
  {start:5,end:11.8,content:"月亮本身不发光。我们看见的月光，是反射的太阳光。"},
  {start:12,end:15.8,content:"太阳照亮月球约一半，我们只能看见其中一部分。"},
  {start:16,end:19.8,content:"月球绕地球运行，我们看见的亮面逐渐增多。"},
  {start:20,end:23.8,content:"满月之后，看见的亮面又逐渐减少。"},
  {start:24,end:27.8,content:"同一轮月亮，于是呈现出不同的月相。"},
  {start:28,end:31.8,content:"月球进入地球的影子时，才是月食。"},
  {start:32,end:34.8,content:"月球轨道有倾角，所以不会每次满月都有月食。"},
  {start:35,end:39.8,content:"一个月相周期，平均约为 29.5 天。"},
  {start:40,end:43.8,content:"月亮没有忽胖忽瘦。改变的是我们看到的亮面。"},
];
const subtitleObjects=SUBTITLES.map(cue=>textAt(cue.content,0,277,23,P.muted));

const film=group(backdrop,...stars,masthead,footer,cover,light,orbital,eclipse,cycle,...subtitleObjects);
film.update((self,time)=>{
  stars.forEach((star,i)=>{star.opacity=.07+.11*(.5+.5*Math.sin(time*.24+i*1.73));});
  cover.opacity=windowOpacity(time,0,5,.8);
  light.opacity=windowOpacity(time,5,12);
  orbital.opacity=windowOpacity(time,12,28);
  eclipse.opacity=windowOpacity(time,28,35);
  cycle.opacity=windowOpacity(time,35,44,.8);
  coverMoon.phase(.45*Math.PI+.12*Math.min(time,5));
  // Keep photon trajectories in the beam's [-274, 143] range, even when seeking.
  photons.forEach((photon,i)=>{const u=((time-5)*.28+i*.31)%1;photon.x=-274+417*((u+1)%1);});
  const angle=orbitalAngle(time);
  position(movingMoon,orbitCenter[0]-orbitRadius*Math.cos(angle),orbitRadius*Math.sin(angle));
  earthView.phase(angle);
  const phaseIndex=phaseNameIndex(angle);
  phaseLabels.forEach((label,i)=>{label.opacity=i===phaseIndex ? 1 : 0;});
  phaseExample.phase(Math.PI/2+.33*Math.sin((time-28)*.65));
  eclipseTint.opacity=.7+.15*smooth((time-29)/3);
  const closingPhase=Math.floor(clamp((time-35.6)/5.6)*8)%8;
  phaseRings.forEach((ring,i)=>{ring.opacity=i===closingPhase ? .74 : .04;});
  subtitleObjects.forEach((object,i)=>{const cue=SUBTITLES[i];object.opacity=cue.visible===false ? 0 : windowOpacity(time,cue.start,cue.end,.28);});
});
add(film);
wait(FILM.duration);
