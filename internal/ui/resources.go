package ui

import "fyne.io/fyne/v2"

var iconSVG = []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256">
  <defs>
    <linearGradient id="g" x1="32" y1="24" x2="224" y2="232" gradientUnits="userSpaceOnUse">
      <stop stop-color="#1FB6A6"/>
      <stop offset="0.55" stop-color="#2E7DD7"/>
      <stop offset="1" stop-color="#5C6BC0"/>
    </linearGradient>
  </defs>
  <rect width="256" height="256" rx="52" fill="#101820"/>
  <path d="M60 92h34l44-38c7-6 18-1 18 9v130c0 10-11 15-18 9l-44-38H60c-9 0-16-7-16-16v-40c0-9 7-16 16-16z" fill="url(#g)"/>
  <path d="M174 91c16 16 16 58 0 74" fill="none" stroke="#F7FFFB" stroke-width="14" stroke-linecap="round"/>
  <path d="M196 70c30 30 30 86 0 116" fill="none" stroke="#F7FFFB" stroke-width="12" stroke-linecap="round" opacity=".86"/>
  <path d="M31 35l190 190" stroke="#C7F6FF" stroke-width="10" stroke-linecap="round" opacity=".25"/>
</svg>`)

func IconResource() fyne.Resource {
	return fyne.NewStaticResource("voicecast.svg", iconSVG)
}
