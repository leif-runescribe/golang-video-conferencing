"use client";

import { useEffect, useRef, useState } from 'react';
import { useParams } from 'next/navigation';

import { Mic, MicOff, Video, VideoOff } from 'lucide-react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

// Remove the need for generateStaticParams by making this a dynamic route
export const dynamic = 'force-dynamic';

export default function Room() {
  const { roomId } = useParams();
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStreams, setRemoteStreams] = useState<MediaStream[]>([]);
  const [isAudioEnabled, setIsAudioEnabled] = useState(true);
  const [isVideoEnabled, setIsVideoEnabled] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null); // Manage error messages
  const localVideoRef = useRef<HTMLVideoElement>(null);
  const ws = useRef<WebSocket | null>(null);
  const peerConnections = useRef<{ [key: string]: RTCPeerConnection }>({});

  useEffect(() => {
    const init = async () => {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          video: true,
          audio: true,
        });
        setLocalStream(stream);
        if (localVideoRef.current) {
          localVideoRef.current.srcObject = stream;
        }

        // Connect to WebSocket server
        ws.current = new WebSocket('ws://localhost:8080/ws');
        ws.current.onmessage = handleWebSocketMessage;
      } catch (error) {
        setErrorMessage('Error accessing media devices. Please check if your camera and microphone are connected.'); // Set error message
      }
    };

    init();

    return () => {
      localStream?.getTracks().forEach(track => track.stop());
      Object.values(peerConnections.current).forEach(pc => pc.close());
      ws.current?.close();
    };
  }, [roomId]);

  const handleWebSocketMessage = async (event: MessageEvent) => {
    const message = JSON.parse(event.data);
    
    if (message.type === 'offer') {
      const pc = createPeerConnection(message.from);
      await pc.setRemoteDescription(new RTCSessionDescription(message.payload));
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      
      ws.current?.send(JSON.stringify({
        type: 'answer',
        roomId,
        from: 'local',
        to: message.from,
        payload: answer,
      }));
    } else if (message.type === 'answer') {
      const pc = peerConnections.current[message.from];
      await pc.setRemoteDescription(new RTCSessionDescription(message.payload));
    } else if (message.type === 'ice-candidate') {
      const pc = peerConnections.current[message.from];
      await pc.addIceCandidate(new RTCIceCandidate(message.payload));
    }
  };

  const createPeerConnection = (peerId: string) => {
    const pc = new RTCPeerConnection({
      iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
    });

    pc.onicecandidate = (event) => {
      if (event.candidate) {
        ws.current?.send(JSON.stringify({
          type: 'ice-candidate',
          roomId,
          from: 'local',
          to: peerId,
          payload: event.candidate,
        }));
      }
    };

    pc.ontrack = (event) => {
      setRemoteStreams(prev => [...prev, event.streams[0]]);
    };

    localStream?.getTracks().forEach(track => {
      pc.addTrack(track, localStream);
    });

    peerConnections.current[peerId] = pc;
    return pc;
  };

  const toggleAudio = () => {
    if (localStream) {
      localStream.getAudioTracks().forEach(track => {
        track.enabled = !isAudioEnabled;
      });
      setIsAudioEnabled(!isAudioEnabled);
    }
  };

  const toggleVideo = () => {
    if (localStream) {
      localStream.getVideoTracks().forEach(track => {
        track.enabled = !isVideoEnabled;
      });
      setIsVideoEnabled(!isVideoEnabled);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 p-4">
      {/* Render Error Message if Media Devices are not found */}
      {errorMessage && (
        <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50">
          <div className="bg-gray-800 p-6 rounded-lg text-white shadow-lg space-y-4">
            <h2 className="text-lg font-bold">Error</h2>
            <p>{errorMessage}</p>
            <Button variant="destructive" onClick={() => setErrorMessage(null)}>
              Close
            </Button>
          </div>
        </div>
      )}

      <div className="max-w-6xl mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <Card className="bg-gray-800 aspect-video relative">
            <video
              ref={localVideoRef}
              autoPlay
              playsInline
              muted
              className="w-full h-full object-cover rounded-lg"
            />
            <div className="absolute bottom-4 left-1/2 transform -translate-x-1/2 flex space-x-2">
              <Button
                size="icon"
                variant={isAudioEnabled ? "default" : "destructive"}
                onClick={toggleAudio}
              >
                {isAudioEnabled ? <Mic /> : <MicOff />}
              </Button>
              <Button
                size="icon"
                variant={isVideoEnabled ? "default" : "destructive"}
                onClick={toggleVideo}
              >
                {isVideoEnabled ? <Video /> : <VideoOff />}
              </Button>
            </div>
          </Card>

          {remoteStreams.map((stream, index) => (
            <Card key={index} className="bg-gray-800 aspect-video">
              <video
                autoPlay
                playsInline
                className="w-full h-full object-cover rounded-lg"
                srcObject={stream}
              />
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}
