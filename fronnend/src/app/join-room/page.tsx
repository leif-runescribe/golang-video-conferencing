"use client";

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useToast } from '@/components/ui/use-toast';
import { Users } from 'lucide-react';

export default function JoinRoom() {
  const [roomId, setRoomId] = useState('');
  const router = useRouter();
  const { toast } = useToast();

  const handleJoinRoom = () => {
    if (!roomId.trim()) {
      toast({
        title: 'Error',
        description: 'Please enter a room ID',
        variant: 'destructive',
      });
      return;
    }

    router.push(`/room/${roomId}`);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-amber-200 via-pink-500 to-purple-500 flex items-center justify-center p-4">
      <Card className="w-full max-w-md p-8 bg-pink-500 text-white">
        <div className="flex flex-col items-center mb-6">
          <Users className="w-12 h-12 text-green-500 mb-4" />
          <h1 className="text-2xl font-bold text-white">Join Room</h1>
        </div>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="roomId" className="text-white">Room ID</Label>
            <Input
              id="roomId"
              placeholder="Enter room ID"
              value={roomId}
              onChange={(e) => setRoomId(e.target.value)}
              className=" text-white"
            />
          </div>

          <Button
            onClick={handleJoinRoom}
            className="w-full"
            size="lg"
            variant="secondary"
          >
            Join Room
          </Button>
        </div>
      </Card>
    </div>
  );
}