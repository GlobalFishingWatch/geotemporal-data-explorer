"use strict";(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[636],{61636:function(t,n,i){function e(t,n,i){return t<n?n:t>i?i:t}i.d(n,{Xg:function(){return M}});const o=Math.log2||function(t){return Math.log(t)*Math.LOG2E};function a(t,n){if(!t)throw new Error(n||"@math.gl/web-mercator: assertion failed.")}const r=Math.PI,u=r/4,h=r/180,s=180/r,m=512,b=85.051129;function f(t){const[n,i]=t;a(Number.isFinite(n)),a(Number.isFinite(i)&&i>=-90&&i<=90,"invalid latitude");const e=i*h;return[m*(n*h+r)/(2*r),m*(r+Math.log(Math.tan(u+.5*e)))/(2*r)]}function c(t){const[n,i]=t,e=n/m*(2*r)-r,o=2*(Math.atan(Math.exp(i/m*(2*r)-r))-u);return[e*s,o*s]}function M(t){const{width:n,height:i,bounds:r,minExtent:u=0,maxZoom:h=24,offset:s=[0,0]}=t,[[m,M],[l,g]]=r,d=function(t=0){if("number"===typeof t)return{top:t,bottom:t,left:t,right:t};return a(Number.isFinite(t.top)&&Number.isFinite(t.bottom)&&Number.isFinite(t.left)&&Number.isFinite(t.right)),t}(t.padding),p=f([m,e(g,-85.051129,b)]),N=f([l,e(M,-85.051129,b)]),F=[Math.max(Math.abs(N[0]-p[0]),u),Math.max(Math.abs(N[1]-p[1]),u)],w=[n-d.left-d.right-2*Math.abs(s[0]),i-d.top-d.bottom-2*Math.abs(s[1])];a(w[0]>0&&w[1]>0);const x=w[0]/F[0],E=w[1]/F[1],k=(d.right-d.left)/2/x,_=(d.bottom-d.top)/2/E,C=c([(N[0]+p[0])/2+k,(N[1]+p[1])/2+_]),I=Math.min(h,o(Math.abs(Math.min(x,E))));return a(Number.isFinite(I)),{longitude:C[0],latitude:C[1],zoom:I}}Math.PI}}]);