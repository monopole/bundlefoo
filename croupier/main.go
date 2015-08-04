// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

// Hacked up version of
// https://godoc.org/golang.org/x/mobile/example/basic

package main

import (
	"encoding/binary"
	"github.com/monopole/mutantfortune/ifc"
	"github.com/monopole/mutantfortune/server/util"
	"github.com/monopole/mutantfortune/service"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/config"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
	"log"
	"v.io/v23"
	_ "v.io/x/ref/runtime/factories/generic"
)

var (
	program  gl.Program
	position gl.Attrib
	offset   gl.Uniform
	color    gl.Uniform
	buf      gl.Buffer

	green float32
	grey  float32
	red   float32
	blue  float32

	touchX float32
	touchY float32
)

func main() {
	app.Main(func(a app.App) {
		doV23 := true
		if doV23 {
			ctx, shutdown := v23.Init()
			defer shutdown()

			// v23.GetNamespace(ctx).SetRoots("/monopole2.mtv.corp.google.com:23000")
			v23.GetNamespace(ctx).SetRoots("/104.197.96.113:3389")
			// A generic server.
			s := util.MakeServer(ctx)

			// Attach the 'fortune service' implementation
			// defined above to a queriable, textual description
			// of the implementation used for service discovery.
			fortune := ifc.FortuneServer(service.Make())

			err := s.Serve(
				"croupier", fortune, util.MakeAuthorizer())
			if err != nil {
				log.Panic("Error serving service: ", err)
			}
		}

		log.Printf("Hi there.\n")

		var c config.Event
		for e := range a.Events() {
			switch e := app.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					onStart()
				case lifecycle.CrossOff:
					onStop()
				}
			case config.Event:
				c = e
				touchX = float32(c.WidthPx / 2)
				touchY = float32(c.HeightPx / 2)
			case paint.Event:
				onPaint(c)
				a.EndPaint(e)
			case touch.Event:
				touchX = e.X
				touchY = e.Y
			}
		}

		// Normal means to end v23 services:
		// <-signals.ShutdownOnSignals(ctx)

	})
}

func onStart() {
	var err error
	program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	buf = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

	position = gl.GetAttribLocation(program, "position")
	color = gl.GetUniformLocation(program, "color")
	offset = gl.GetUniformLocation(program, "offset")

	// TODO(crawshaw): the debug package needs to put GL state init here
	// Can this be an app.RegisterFilter call now??
}

func onStop() {
	gl.DeleteProgram(program)
	gl.DeleteBuffer(buf)
}

func onPaint(c config.Event) {
	gl.ClearColor(1, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(program)

	green += 0.01
	if green > 1 {
		green = 0
	}
	gl.Uniform4f(color, 0, green, 0, 1)

	gl.Uniform2f(offset, touchX/float32(c.WidthPx), touchY/float32(c.HeightPx))

	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.EnableVertexAttribArray(position)
	gl.VertexAttribPointer(position, coordsPerVertex, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, vertexCount)
	gl.DisableVertexAttribArray(position)

	debug.DrawFPS(c)
}

var triangleData = f32.Bytes(binary.LittleEndian,
	0.0, 0.4, 0.0, // top left
	0.0, 0.0, 0.0, // bottom left
	0.4, 0.0, 0.0, // bottom right
)

const (
	coordsPerVertex = 3
	vertexCount     = 3
)

const vertexShader = `#version 100
uniform vec2 offset;
attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`
